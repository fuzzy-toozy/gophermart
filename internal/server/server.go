package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fuzzy-toozy/gophermart/internal/config"
	"github.com/fuzzy-toozy/gophermart/internal/controllers"
	"github.com/fuzzy-toozy/gophermart/internal/services"

	"github.com/fuzzy-toozy/gophermart/internal/database"
	"github.com/fuzzy-toozy/gophermart/internal/database/repo"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	logger            *zap.SugaredLogger
	config            *config.Config
	stopCtx           context.Context
	stopFunc          context.CancelFunc
	serviceStorage    *database.ServiceStorage
	userConroller     *controllers.UserController
	balanceController *controllers.BalanceContoller
	ordersController  *controllers.OrderController
	processController *controllers.ProcessContoller
	processService    *services.ProcessingService
	router            *gin.Engine
	httpServer        *http.Server
}

func (s *Server) runProcessor(ctx context.Context) {
	t := time.NewTicker(s.config.ProcessingInteval)

	for {
		select {
		case <-t.C:
			err := s.processService.ProcessOrders(ctx)
			if err != nil {
				s.logger.Errorf("Failed to process orders: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) Run() {
	start := func() error {
		return s.httpServer.ListenAndServe()
	}

	stop := func() error {
		err := s.httpServer.Shutdown(context.Background())
		if err != nil {
			err = fmt.Errorf("server shutdown failed: %w", err)
		}

		if err := s.serviceStorage.Close(); err != nil {
			s.logger.Errorf("Failed to close storage: %v", err)
		}

		return err
	}

	ctx := context.Background()
	stopCtx, cancel := context.WithCancel(ctx)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		<-c
		cancel()
	}()

	g, gCtx := errgroup.WithContext(stopCtx)

	g.Go(func() error {
		return start()
	})

	g.Go(func() error {
		s.runProcessor(gCtx)
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		return stop()
	})

	err := g.Wait()

	if err != nil {
		s.logger.Infof("Server exit reason: %v", err)
	}
}

func (s *Server) setupRouting() {
	s.router.POST("/api/user/register", s.userConroller.Register)
	s.router.POST("/api/user/login", s.userConroller.Login)

	authGrp := s.router.Group("/api/user")
	authGrp.Use(s.userConroller.Authenticate)
	{
		authGrp.GET("/orders", s.ordersController.GetAllOrders)
		authGrp.GET("/withdrawals", s.balanceController.GetWithdrawals)
		authGrp.GET("/balance", s.balanceController.GetBalanceData)

		authGrp.POST("/orders", s.ordersController.AddNewOrder)
		authGrp.POST("/balance/withdraw", s.processController.Withdraw)
	}
}

func NewServer(c *config.Config, l AppLogger) (*Server, error) {
	err := database.Migrate(c.DatabaseConfig.ConnURI)
	if err != nil {
		return nil, fmt.Errorf("failed db migration: %v", err)
	}

	gin.DefaultWriter = io.MultiWriter(l.LogFile, os.Stdout)

	serviceStorage, err := database.NewServiceStorage(c.DatabaseConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to setup repository: %v", err)
	}

	userRepo := repo.NewUserServiceRepo(*serviceStorage)
	userService := services.NewUserService(userRepo, services.NewTokenService(c.SecretKey, c.TokenLifetime))
	userController := controllers.NewUserController(userService, l.Logger)

	orderRepo := repo.NewOrderServiceRepo(serviceStorage)
	orderService := services.NewOrderService(orderRepo)
	orderController := controllers.NewOrderController(orderService, l.Logger)

	balanceRepo := repo.NewBalanceServiceRepo(serviceStorage)
	balanceService := services.NewBalanceService(balanceRepo)
	balanceController := controllers.NewBalanceController(balanceService, l.Logger)

	processRepo := repo.NewProcessRepo(serviceStorage, balanceRepo, orderRepo)
	accrualService := services.NewAccrualService(&http.Client{}, c.AccrualAddress, l.Logger)
	processService := services.NewProcessingService(processRepo, accrualService, l.Logger)
	procesController := controllers.NewProcessController(processService, l.Logger)

	c.ServerAddress = strings.TrimPrefix(c.ServerAddress, "http://")

	s := Server{
		config:            c,
		logger:            l.Logger,
		serviceStorage:    serviceStorage,
		userConroller:     userController,
		ordersController:  orderController,
		balanceController: balanceController,
		processController: procesController,
		processService:    processService,
		router:            gin.Default(),
	}

	httpServer := http.Server{
		Addr:         c.ServerAddress,
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
		IdleTimeout:  c.IdleTimeout,
		Handler:      s.router,
	}

	s.setupRouting()

	s.httpServer = &httpServer

	return &s, nil
}

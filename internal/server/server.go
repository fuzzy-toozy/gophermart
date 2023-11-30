package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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
			s.logger.Debugf("Started processing orders")
			err := s.processService.ProcessOrders(ctx)
			if err != nil {
				s.logger.Errorf("Failed to process orders: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) Run() error {
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

	return g.Wait()
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
		authGrp.POST("/balance/widthdraw", s.processController.Withdraw)
	}
}

func NewServer(logger *zap.SugaredLogger) (*Server, error) {
	c, err := config.BuildConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %v", err)
	}

	err = database.Migrate(c.DatabaseConfig.ConnURI)
	if err != nil {
		return nil, fmt.Errorf("failed db migration: %v", err)
	}

	serviceStorage, err := database.NewServiceStorage(c.DatabaseConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to setup repository: %v", err)
	}

	userRepo := repo.NewUserServiceRepo(*serviceStorage)
	userService := services.NewUserService(userRepo, services.NewTokenService(c.SecretKey, c.TokenLifetime))
	userController := controllers.NewUserController(userService, logger)

	orderRepo := repo.NewOrderServiceRepo(serviceStorage)
	orderService := services.NewOrderService(orderRepo)
	orderController := controllers.NewOrderController(orderService, logger)

	balanceRepo := repo.NewBalanceServiceRepo(serviceStorage)
	balanceService := services.NewBalanceService(balanceRepo)
	balanceController := controllers.NewBalanceController(balanceService, logger)

	processRepo := repo.NewProcessRepo(serviceStorage, balanceRepo, orderRepo)
	accrualService := services.NewAccrualService(&http.Client{}, c.AccrualAddress)
	processService := services.NewProcessingService(processRepo, accrualService, logger)
	procesController := controllers.NewProcessController(processService, logger)

	s := Server{
		config:            c,
		logger:            logger,
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
		ReadTimeout:  c.ReadTimeoutSec,
		WriteTimeout: c.WriteTimeoutSec,
		IdleTimeout:  c.IdleTimeoutSec,
		Handler:      s.router,
	}

	s.setupRouting()

	s.httpServer = &httpServer

	return &s, nil
}
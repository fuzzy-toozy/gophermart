package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/fuzzy-toozy/gophermart/internal/common"
	"github.com/fuzzy-toozy/gophermart/internal/models"
	"github.com/fuzzy-toozy/gophermart/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	tokenCookieKey = "Auth"
)

type UserController struct {
	service *services.UserService
	logger  *zap.SugaredLogger
}

func NewUserController(service *services.UserService, logger *zap.SugaredLogger) *UserController {
	return &UserController{
		service: service,
		logger:  logger,
	}
}

func (c *UserController) setCookie(ctx *gin.Context, token string) {
	ctx.SetCookie(tokenCookieKey, token, int(c.service.AuthDuration()/time.Second), "/",
		strings.Split(ctx.Request.Host, ":")[0], false, true)
}

func (c *UserController) Register(ctx *gin.Context) {
	user := models.User{}

	if err := ctx.BindJSON(&user); err != nil {
		c.logger.Debugf("Failed to bind request data: %v", err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	token, err := c.service.Register(ctx, &user)
	if err != nil {
		c.logger.Debugf("Failed to register new user: %v", err)
		ctx.AbortWithStatus(err.GetStatus())
		return
	}

	c.setCookie(ctx, token)
	ctx.Status(http.StatusOK)
}

func (c *UserController) Login(ctx *gin.Context) {
	user := models.User{}

	if err := ctx.BindJSON(&user); err != nil {
		c.logger.Debugf("Failed to bind request data: %v", err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	token, err := c.service.Login(ctx, &user)
	if err != nil {
		c.logger.Debugf("Failed to login user '%v': %v", user.Username, err)
		ctx.AbortWithStatus(err.GetStatus())
		return
	}

	c.setCookie(ctx, token)
	ctx.Status(http.StatusOK)
}

func (c *UserController) Authenticate(ctx *gin.Context) {
	signedToken, err := ctx.Cookie(tokenCookieKey)
	if err != nil {
		c.logger.Debug("Authentication failed. No jwt token provided")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	username, err := c.service.Authenticate(signedToken)
	if err != nil {
		c.logger.Debugf("Authentication failed: %v", err)
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx.Set(common.UsernameCtxKey, username)
	ctx.Next()
}

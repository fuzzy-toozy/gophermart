package controllers

import (
	"net/http"

	"github.com/fuzzy-toozy/gophermart/internal/common"
	"github.com/fuzzy-toozy/gophermart/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BalanceContoller struct {
	service *services.BalanceService
	logger  *zap.SugaredLogger
}

func NewBalanceController(service *services.BalanceService, logger *zap.SugaredLogger) *BalanceContoller {
	return &BalanceContoller{
		service: service,
		logger:  logger,
	}
}

func (c *BalanceContoller) GetBalanceData(ctx *gin.Context) {
	username := ctx.GetString(common.UsernameCtxKey)

	balance, err := c.service.GetUserBalance(ctx, username)
	if err != nil {
		c.logger.Debugf("failed to get balance data for user %v: %v", username, err)
		ctx.AbortWithStatus(err.GetStatus())
		return
	}

	ctx.JSON(http.StatusOK, balance)
}

func (c *BalanceContoller) GetWithdrawals(ctx *gin.Context) {
	username := ctx.GetString(common.UsernameCtxKey)

	wd, err := c.service.GetAllUserWithdrawals(ctx, username)
	if err != nil {
		c.logger.Debugf("Failed to get all withdrawals for user %v: %v", username, err)
		ctx.AbortWithStatus(err.GetStatus())
		return
	}

	ctx.JSON(http.StatusOK, wd)
}

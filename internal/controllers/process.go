package controllers

import (
	"net/http"

	"github.com/fuzzy-toozy/gophermart/internal/common"
	"github.com/fuzzy-toozy/gophermart/internal/models"
	"github.com/fuzzy-toozy/gophermart/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProcessContoller struct {
	service *services.ProcessingService
	logger  *zap.SugaredLogger
}

func NewProcessController(service *services.ProcessingService, logger *zap.SugaredLogger) *ProcessContoller {
	return &ProcessContoller{
		service: service,
		logger:  logger,
	}
}

func (c *ProcessContoller) Withdraw(ctx *gin.Context) {
	wd := models.Withdraw{}

	if err := ctx.BindJSON(&wd); err != nil {
		c.logger.Debugf("Failed to bind request data: %v", err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	username := ctx.GetString(common.UsernameCtxKey)

	err := c.service.Withdraw(ctx, &wd, username)

	if err != nil {
		c.logger.Debugf("Failed to withdraw balance for user '%v': %v", username, err)
		ctx.AbortWithStatus(err.GetStatus())
		return
	}

	ctx.Status(http.StatusOK)
}

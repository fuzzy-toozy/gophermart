package controllers

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/fuzzy-toozy/gophermart/internal/common"
	"github.com/fuzzy-toozy/gophermart/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OrderController struct {
	service *services.OrderService
	logger  *zap.SugaredLogger
}

func NewOrderController(service *services.OrderService, logger *zap.SugaredLogger) *OrderController {
	return &OrderController{
		service: service,
		logger:  logger,
	}
}

func (c *OrderController) AddNewOrder(ctx *gin.Context) {
	buf := new(bytes.Buffer)

	_, err := buf.ReadFrom(ctx.Request.Body)
	if err != nil && !errors.Is(err, io.EOF) {
		c.logger.Debugf("Failed to read request body: %v", err)
		return
	}

	orderNumber := buf.String()
	username := ctx.GetString(common.UsernameCtxKey)

	serr := c.service.CheckOrderNumber(ctx, username, orderNumber)
	if serr != nil {
		c.logger.Debugf("Checking order number '%v' for user '%v' result: %v", username, orderNumber, serr)
		ctx.AbortWithStatus(serr.GetStatus())
		return
	}

	serr = c.service.AddNewOrder(ctx, username, orderNumber)
	if serr != nil {
		c.logger.Debugf("Adding new order failed for user '%v': %v", username, serr)
		ctx.AbortWithStatus(serr.GetStatus())
		return
	}

	ctx.Status(http.StatusAccepted)
}

func (c *OrderController) GetAllOrders(ctx *gin.Context) {
	username := ctx.GetString(common.UsernameCtxKey)
	results, err := c.service.GetAllOrders(ctx, username)
	if err != nil {
		c.logger.Debugf("Could't get all orders for user '%v': %v", username, err)
		ctx.AbortWithStatus(err.GetStatus())
		return
	}

	ctx.JSON(http.StatusOK, results)
}

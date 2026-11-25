package handler

import (
	"errors"
	"net/http"

	"example.com/train-ticket-system/backend-app/internal/repository"
	"example.com/train-ticket-system/backend-app/internal/service"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(
	s *service.OrderService,
) *OrderHandler {
	return &OrderHandler{service: s}
}

// RegisterRoutes 注册订单相关路由。
func (h *OrderHandler) RegisterRoutes(router *gin.Engine) {

	group := router.Group("/api/orders")

	group.GET("", h.list)
	group.GET("/:order_no", h.get)
}

// GET /api/orders/:order_no
func (h *OrderHandler) get(c *gin.Context) {

	order, err := h.service.Get(
		c.Request.Context(),
		c.Param("order_no"),
	)

	if errors.Is(err, repository.ErrOrderNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "订单不存在",
		})
		return
	}

	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": order,
	})
}

// GET /api/orders?user_id=U001
func (h *OrderHandler) list(c *gin.Context) {

	orders, err := h.service.ListByUser(
		c.Request.Context(),
		c.Query("user_id"),
	)

	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": orders,
	})
}

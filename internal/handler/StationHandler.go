package handler

import (
	"net/http"
	"strconv"

	"example.com/train-ticket-system/backend-app/internal/service"
	"github.com/gin-gonic/gin"
)

// StationHandler 处理站点查询相关请求。
type StationHandler struct {
	service *service.TrainService
}

// NewStationHandler 创建站点 Handler。
func NewStationHandler(
	s *service.TrainService,
) *StationHandler {
	return &StationHandler{
		service: s,
	}
}

// RegisterRoutes 注册站点相关路由。
func (h *StationHandler) RegisterRoutes(
	router *gin.Engine,
) {
	group := router.Group("/api/stations")

	group.GET("/suggest", h.suggest)
}

// GET /api/stations/suggest?prefix=成都&limit=10
func (h *StationHandler) suggest(c *gin.Context) {
	// 1. 获取查询参数。
	prefix := c.Query("prefix")
	limitStr := c.DefaultQuery("limit", "10")

	// 2. 将 limit 转换为整数。
	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit <= 0 || limit > 50 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "limit 必须是 1 到 50 之间的整数",
		})
		return
	}

	// 3. 调用 Service 查询站点。
	stations, err := h.service.SuggestStations(
		c.Request.Context(),
		prefix,
		limit,
	)

	if err != nil {
		respondError(c, err)
		return
	}

	// 4. 返回查询结果。
	c.JSON(http.StatusOK, gin.H{
		"data": stations,
	})
}

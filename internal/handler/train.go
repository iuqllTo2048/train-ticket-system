package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"example.com/train-ticket-system/backend-app/internal/repository"
	"example.com/train-ticket-system/backend-app/internal/service"
	"example.com/train-ticket-system/backend-app/internal/taskmanager"
	"github.com/gin-gonic/gin"
)

type TrainHandler struct {
	service     *service.TrainService
	taskManager *taskmanager.TaskManager
}

func NewTrainHandler(
	s *service.TrainService,
	tm *taskmanager.TaskManager,
) *TrainHandler {
	return &TrainHandler{
		service:     s,
		taskManager: tm,
	}
}

func (h *TrainHandler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/api/trains")

	group.GET("", h.list)
	group.GET("/:id", h.get)

	group.POST("", h.create)
	group.PATCH("/:id", h.update)
	group.DELETE("/:id", h.delete)

	group.POST("/:id/purchase", h.purchase)
	group.POST("/:id/purchase/async", h.purchaseAsync)
}

type createTrainRequest struct {
	ID string `json:"id" binding:"required"`

	FromStation string `json:"from_station" binding:"required"`
	ToStation   string `json:"to_station" binding:"required"`

	PriceCents int64 `json:"price_cents" binding:"required"`
	TotalSeats int   `json:"total_seats" binding:"required"`
}

type patchTrainRequest struct {
	FromStation *string `json:"from_station"`
	ToStation   *string `json:"to_station"`
	PriceCents  *int64  `json:"price_cents"`
}

type purchaseRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Quantity int    `json:"quantity" binding:"required"`
}

// GET /api/trains
func (h *TrainHandler) list(c *gin.Context) {
	trains, err := h.service.List(
		c.Request.Context(),
		c.Query("from"),
		c.Query("to"),
	)

	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": trains,
	})
}

// GET /api/trains/:id
func (h *TrainHandler) get(c *gin.Context) {
	train, err := h.service.Get(
		c.Request.Context(),
		c.Param("id"),
	)

	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": train,
	})
}

// POST /api/trains
func (h *TrainHandler) create(c *gin.Context) {
	var req createTrainRequest

	//if err := c.ShouldBindJSON(&req); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": "JSON 格式或必填字段不正确",
	//	})
	//	return
	//}
	//执行json解析
	bindErr := c.ShouldBindJSON(&req)
	//输出实际执行结果
	fmt.Printf("DEBUG bindErr = %v\n", bindErr)
	fmt.Printf("DEBUG req = %+v\n", req)
	fmt.Printf("DEBUG req.ID = %q\n", req.ID)
	//判断解析是否出错
	if bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "JSON 格式或必填字段不正确",
		})
		return
	}

	train, err := h.service.Create(
		c.Request.Context(),
		service.CreateTrainInput{
			ID:          req.ID,
			FromStation: req.FromStation,
			ToStation:   req.ToStation,
			PriceCents:  req.PriceCents,
			TotalSeats:  req.TotalSeats,
		},
	)

	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": train,
	})
}

// PATCH /api/trains/:id
func (h *TrainHandler) update(c *gin.Context) {
	var req patchTrainRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "JSON 格式不正确",
		})
		return
	}

	train, err := h.service.Update(
		c.Request.Context(),
		c.Param("id"),
		service.TrainPatch{
			FromStation: req.FromStation,
			ToStation:   req.ToStation,
			PriceCents:  req.PriceCents,
		},
	)

	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": train,
	})
}

// DELETE /api/trains/:id
func (h *TrainHandler) delete(c *gin.Context) {
	err := h.service.Delete(
		c.Request.Context(),
		c.Param("id"),
	)

	if err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// POST /api/trains/:id/purchase
func (h *TrainHandler) purchase(c *gin.Context) {
	var req purchaseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "JSON 格式或购票参数不正确",
		})
		return
	}

	order, err := h.service.Purchase(
		c.Request.Context(),
		c.Param("id"),
		req.UserID,
		req.Quantity,
	)

	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "购票成功",
		"data":    order,
	})
}

// POST /api/trains/:id/purchase/async
func (h *TrainHandler) purchaseAsync(c *gin.Context) {

	// 1. 解析购票请求参数。
	var req purchaseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "JSON 格式或购票参数不正确",
		})
		return
	}

	// 2. 提前提取业务参数。
	trainID := c.Param("id")
	userID := req.UserID
	quantity := req.Quantity

	// 3. 将购票业务包装为异步任务。
	taskID, err := h.taskManager.SubmitWithResult(
		"purchase",
		func() (string, error) {

			// 创建独立的任务执行 Context。
			ctx, cancel := context.WithTimeout(
				context.Background(),
				10*time.Second,
			)
			defer cancel()

			// 复用原有的真实购票业务。
			order, err := h.service.Purchase(
				ctx,
				trainID,
				userID,
				quantity,
			)

			if err != nil {
				return "", err
			}

			// 返回正式订单编号。
			return order.OrderNo, nil
		},
	)

	// 4. 处理任务提交失败。
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "购票任务提交失败",
			"task_id": taskID,
		})
		return
	}

	// 5. 返回任务提交成功响应。
	c.JSON(http.StatusAccepted, gin.H{
		"message": "购票任务已提交",
		"task_id": taskID,
	})
}

// 统一处理业务错误。
func respondError(c *gin.Context, err error) {
	switch {

	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "车次不存在",
		})

	case errors.Is(err, repository.ErrInsufficientSeats):
		c.JSON(http.StatusConflict, gin.H{
			"error": "余票不足",
		})

	case errors.Is(err, repository.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{
			"error": "车次已存在，或有关联的购票记录",
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "服务器内部错误",
		})
	}
}

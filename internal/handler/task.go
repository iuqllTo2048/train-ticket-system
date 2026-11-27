package handler

import (
	"net/http"
	"strconv"

	"example.com/train-ticket-system/backend-app/internal/taskmanager"
	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	manager *taskmanager.TaskManager
}

// NewTaskHandler 创建任务查询 Handler。
func NewTaskHandler(
	manager *taskmanager.TaskManager,
) *TaskHandler {

	return &TaskHandler{
		manager: manager,
	}
}

// RegisterRoutes 注册任务查询接口。
func (h *TaskHandler) RegisterRoutes(router *gin.Engine) {

	group := router.Group("/api/tasks")

	group.GET("", h.list)
	group.GET("/:id", h.get)
}

// GET /api/tasks/:id
func (h *TaskHandler) get(c *gin.Context) {

	// 1. 将路径参数转换为 uint64。
	id, err := strconv.ParseUint(
		c.Param("id"),
		10,
		64,
	)

	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "任务 ID 不合法",
		})
		return
	}

	// 2. 查询任务信息。
	task, ok := h.manager.GetTask(id)

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "任务不存在",
		})
		return
	}

	// 3. 返回任务信息。
	c.JSON(http.StatusOK, gin.H{
		"data": task,
	})
}

// GET /api/tasks
func (h *TaskHandler) list(c *gin.Context) {

	tasks := h.manager.ListTasks()

	c.JSON(http.StatusOK, gin.H{
		"data": tasks,
	})
}

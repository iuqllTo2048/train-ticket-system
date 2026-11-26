package taskmanager

import "time"

// TaskStatus 表示任务的生命周期状态。
type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskRunning   TaskStatus = "running"
	TaskCompleted TaskStatus = "completed"
	TaskFailed    TaskStatus = "failed"
)

// TaskFunc 表示可以返回业务错误的任务函数。
type TaskFunc func() error

// TaskInfo 保存任务的元数据及执行状态。
type TaskInfo struct {
	ID         uint64        `json:"id"`
	Name       string        `json:"name"`
	Status     TaskStatus    `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	StartedAt  time.Time     `json:"started_at"`
	FinishedAt time.Time     `json:"finished_at"`
	Duration   time.Duration `json:"duration"`
	Error      string        `json:"error"`
}

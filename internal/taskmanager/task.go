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

// TaskFunc 表示只返回业务错误的任务函数。
type TaskFunc func() error

// ResultTaskFunc 表示能够返回业务结果引用的任务函数。
type ResultTaskFunc func() (string, error)

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

	// 保存业务执行结果的引用，例如订单编号。
	ResultRef string `json:"result_ref,omitempty"`
}

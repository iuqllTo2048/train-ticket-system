package taskmanager

import (
	"fmt"
	"sync"
	"time"

	"example.com/train-ticket-system/backend-app/internal/workerpool"
)

type TaskManager struct {
	pool *workerpool.WorkerPool

	mu     sync.RWMutex
	tasks  map[uint64]*TaskInfo
	nextID uint64
}

// NewTaskManager 创建任务管理器。
func NewTaskManager(pool *workerpool.WorkerPool) *TaskManager {
	if pool == nil {
		panic("worker pool must not be nil")
	}

	return &TaskManager{
		pool:  pool,
		tasks: make(map[uint64]*TaskInfo),
	}
}

// createTask 创建任务记录并返回唯一任务 ID。
func (m *TaskManager) createTask(name string) uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextID++
	id := m.nextID

	m.tasks[id] = &TaskInfo{
		ID:        id,
		Name:      name,
		Status:    TaskPending,
		CreatedAt: time.Now(),
	}

	return id
}

// GetTask 根据任务 ID 查询任务信息。
func (m *TaskManager) GetTask(id uint64) (TaskInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, ok := m.tasks[id]
	if !ok {
		return TaskInfo{}, false
	}

	return *info, true
}

// ListTasks 获取全部任务的信息快照。
func (m *TaskManager) ListTasks() []TaskInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]TaskInfo, 0, len(m.tasks))

	for _, info := range m.tasks {
		result = append(result, *info)
	}

	return result
}

// markRunning 将任务标记为执行中。
func (m *TaskManager) markRunning(id uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, ok := m.tasks[id]
	if !ok || info.Status != TaskPending {
		return
	}

	info.Status = TaskRunning
	info.StartedAt = time.Now()
}

// markCompleted 将任务标记为执行成功。
func (m *TaskManager) markCompleted(id uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, ok := m.tasks[id]
	if !ok || info.Status != TaskRunning {
		return
	}

	info.Status = TaskCompleted
	info.FinishedAt = time.Now()
	info.Duration = info.FinishedAt.Sub(info.StartedAt)
}

// markFailed 将任务标记为执行失败。
func (m *TaskManager) markFailed(id uint64, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, ok := m.tasks[id]
	if !ok {
		return
	}

	if info.Status != TaskPending && info.Status != TaskRunning {
		return
	}

	info.Status = TaskFailed
	info.FinishedAt = time.Now()

	if !info.StartedAt.IsZero() {
		info.Duration = info.FinishedAt.Sub(info.StartedAt)
	}

	if err != nil {
		info.Error = err.Error()
	}
}

// Submit 创建任务记录，并将任务提交给 WorkerPool。
func (m *TaskManager) Submit(
	name string,
	task TaskFunc,
) (uint64, error) {

	if task == nil {
		return 0, fmt.Errorf("task must not be nil")
	}

	// 1. 创建任务记录
	id := m.createTask(name)

	// 2. 包装原始业务函数
	wrappedTask := func() {

		// 捕获业务任务产生的 panic
		defer func() {
			if recovered := recover(); recovered != nil {
				m.markFailed(
					id,
					fmt.Errorf("task panic: %v", recovered),
				)
			}
		}()

		// 记录任务开始执行
		m.markRunning(id)

		// 执行原始业务函数
		err := task()

		// 处理业务错误
		if err != nil {
			m.markFailed(id, err)
			return
		}

		// 记录任务执行成功
		m.markCompleted(id)
	}

	// 3. 将包装后的函数提交给 WorkerPool
	if !m.pool.Submit(wrappedTask) {
		err := fmt.Errorf("worker pool rejected task")
		m.markFailed(id, err)
		return id, err
	}

	// 4. 返回任务 ID
	return id, nil
}

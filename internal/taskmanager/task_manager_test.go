package taskmanager

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"example.com/train-ticket-system/backend-app/internal/workerpool"
)

// 测试任务的完整生命周期。
func TestTaskLifecycle(t *testing.T) {
	pool := workerpool.NewWorkerPool(1, 4)
	manager := NewTaskManager(pool)

	started := make(chan struct{})
	release := make(chan struct{})

	var once sync.Once

	releaseTask := func() {
		once.Do(func() {
			close(release)
		})
	}

	// 即使测试失败，也释放任务并关闭协程池。
	defer pool.Shutdown()
	defer releaseTask()

	// 提交第一个任务，让它占用唯一的 Worker。
	id1, err := manager.Submit("任务1", func() error {
		close(started)
		<-release
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	// 等待任务1真正开始执行。
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("任务1启动超时")
	}

	// 验证任务1处于 Running。
	info1, ok := manager.GetTask(id1)

	if !ok || info1.Status != TaskRunning {
		t.Fatal("任务1应该处于 Running 状态")
	}

	// 提交第二个任务。
	id2, err := manager.Submit("任务2", func() error {
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	// 唯一的 Worker 被占用，任务2必然处于 Pending。
	info2, ok := manager.GetTask(id2)

	if !ok || info2.Status != TaskPending {
		t.Fatal("任务2应该处于 Pending 状态")
	}

	t.Log("任务1：Running")
	t.Log("任务2：Pending")

	// 验证返回的是副本。
	info2.Status = TaskFailed

	actual, _ := manager.GetTask(id2)

	if actual.Status != TaskPending {
		t.Fatal("外部修改副本影响了原始任务记录")
	}

	// 释放任务1，让 Worker 继续消费任务2。
	releaseTask()

	// 等待所有任务执行结束。
	pool.Shutdown()

	// 检查两个任务的最终状态。
	tasks := manager.ListTasks()

	if len(tasks) != 2 {
		t.Fatalf("预期2个任务，实际%d个", len(tasks))
	}

	for _, info := range tasks {
		if info.Status != TaskCompleted {
			t.Fatalf("任务%d没有正常完成", info.ID)
		}

		if info.CreatedAt.IsZero() ||
			info.StartedAt.IsZero() ||
			info.FinishedAt.IsZero() {
			t.Fatal("任务时间记录不完整")
		}

		if info.StartedAt.Before(info.CreatedAt) ||
			info.FinishedAt.Before(info.StartedAt) {
			t.Fatal("任务时间顺序错误")
		}

		if info.Duration != info.FinishedAt.Sub(info.StartedAt) {
			t.Fatal("任务耗时计算错误")
		}

		t.Logf(
			"任务%d：%s，执行耗时：%v",
			info.ID,
			info.Status,
			info.Duration,
		)
	}
}

// 测试业务错误、panic 以及 Worker 的持续执行能力。
func TestTaskFailure(t *testing.T) {
	pool := workerpool.NewWorkerPool(1, 4)
	manager := NewTaskManager(pool)

	defer pool.Shutdown()

	// 任务1：返回业务错误。
	errorID, err := manager.Submit("业务失败任务", func() error {
		return errors.New("余票不足")
	})

	if err != nil {
		t.Fatal(err)
	}

	// 任务2：触发 panic。
	panicID, err := manager.Submit("异常任务", func() error {
		panic("模拟系统异常")
	})

	if err != nil {
		t.Fatal(err)
	}

	// 任务3：正常执行。
	successID, err := manager.Submit("正常任务", func() error {
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	pool.Shutdown()

	errorInfo, _ := manager.GetTask(errorID)
	panicInfo, _ := manager.GetTask(panicID)
	successInfo, _ := manager.GetTask(successID)

	if errorInfo.Status != TaskFailed ||
		errorInfo.Error != "余票不足" {
		t.Fatal("业务错误状态记录失败")
	}

	if panicInfo.Status != TaskFailed ||
		!strings.Contains(panicInfo.Error, "模拟系统异常") {
		t.Fatal("panic 状态记录失败")
	}

	if successInfo.Status != TaskCompleted {
		t.Fatal("Worker 未能继续执行后续任务")
	}

	t.Logf("业务错误：%s", errorInfo.Error)
	t.Logf("panic 错误：%s", panicInfo.Error)
	t.Log("后续任务正常完成，Worker 未退出")
}

// 测试任务提交失败。
func TestTaskRejected(t *testing.T) {
	pool := workerpool.NewWorkerPool(1, 4)
	manager := NewTaskManager(pool)

	pool.Shutdown()

	// nil 任务不应该创建任务记录。
	nilID, err := manager.Submit("空任务", nil)

	if nilID != 0 || err == nil {
		t.Fatal("nil 任务应该被拒绝")
	}

	if len(manager.ListTasks()) != 0 {
		t.Fatal("nil 任务不应该生成记录")
	}

	// 协程池关闭后提交任务。
	id, err := manager.Submit("拒绝任务", func() error {
		return nil
	})

	if err == nil {
		t.Fatal("关闭后的协程池不应该接收任务")
	}

	info, ok := manager.GetTask(id)

	if !ok || info.Status != TaskFailed {
		t.Fatal("被拒绝的任务应该记录为 Failed")
	}

	if !info.StartedAt.IsZero() {
		t.Fatal("被拒绝的任务不应该有开始时间")
	}

	if info.FinishedAt.IsZero() || info.Duration != 0 {
		t.Fatal("被拒绝任务的时间记录错误")
	}

	if info.Error == "" {
		t.Fatal("应该记录任务被拒绝的原因")
	}

	t.Logf("任务提交被拒绝：%s", info.Error)
}

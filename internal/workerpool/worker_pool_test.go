package workerpool

import "testing"

func TestWorkerPoolConcurrencyAndShutdown(t *testing.T) {
	pool := NewWorkerPool(2, 1)

	started := make(chan struct{}, 3)
	release := make(chan struct{})
	finished := make(chan struct{}, 3)

	task := func() {
		started <- struct{}{}
		<-release
		finished <- struct{}{}
	}

	// 提交 A、B，并等待两个 Worker 都进入任务函数。
	if !pool.Submit(task) || !pool.Submit(task) {
		t.Fatal("A、B 提交失败")
	}
	<-started
	<-started

	// 两个 Worker 都在等 release，C 此时占据唯一的排队位置。
	if !pool.Submit(task) {
		t.Fatal("C 提交失败")
	}

	shutdownDone := make(chan struct{})
	go func() {
		pool.Shutdown()
		close(shutdownDone)
	}()

	// 队列满时，这次提交会等到队列关闭并被拒绝。
	// 如果 Shutdown 已经关闭了队列，这次提交会直接被拒绝。
	if pool.Submit(func() {}) {
		t.Error("关闭时接受了额外的任务")
	}

	select {
	case <-started:
		t.Error("两个 Worker 都忙时，第三个任务已经开始执行")
	default:
	}

	select {
	case <-shutdownDone:
		t.Error("任务尚未放行，Shutdown 就已经返回")
	default:
	}

	// 关闭通道后，A、B 都能继续，之后取到 C 的 Worker 也能继续。
	close(release)
	<-shutdownDone

	if got := len(finished); got != 3 {
		t.Errorf("Shutdown 返回时完成了 %d 个任务，期望 3 个", got)
	}
}

func TestWorkerPoolContinuesAfterPanic(t *testing.T) {
	pool := NewWorkerPool(1, 2)
	ranAfterPanic := make(chan struct{})

	if !pool.Submit(func() {
		panic("测试任务触发 panic")
	}) {
		t.Fatal("第一个任务提交失败")
	}

	if !pool.Submit(func() {
		close(ranAfterPanic)
	}) {
		t.Fatal("第二个任务提交失败")
	}

	pool.Shutdown()

	select {
	case <-ranAfterPanic:
		// 池里只有一个 Worker，说明它在处理 panic 后继续取了任务。
	default:
		t.Fatal("第一个任务触发 panic 后，第二个任务没有执行")
	}
}

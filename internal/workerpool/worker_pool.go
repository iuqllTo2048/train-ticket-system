package workerpool

import (
	"log"
	"runtime/debug"
	"sync"

	"example.com/train-ticket-system/backend-app/internal/datastructure"
)

// Task 表示一个可以被 Worker 执行的任务。
type Task func()

type WorkerPool struct {
	queue   *datastructure.BlockingQueue[Task]
	workers sync.WaitGroup
}

// NewWorkerPool 创建工作协程池并立即启动 Worker。
func NewWorkerPool(workerCount, queueCapacity int) *WorkerPool {
	if workerCount <= 0 {
		panic("worker count must be greater than 0")
	}

	pool := &WorkerPool{
		queue: datastructure.NewBlockingQueue[Task](queueCapacity),
	}

	pool.workers.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go pool.worker()
	}

	return pool
}

// Submit 提交任务。队列满时等待，任务成功入队时返回 true。
// 任务为 nil 或队列已经关闭时返回 false。
func (p *WorkerPool) Submit(task Task) bool {
	if task == nil {
		return false
	}

	return p.queue.Enqueue(task)
}

// Shutdown 停止接收新任务，并等待所有已接收的任务执行结束。
// 应由池的使用方调用。Worker 正在执行的任务不能调用同一个池的 Shutdown。
func (p *WorkerPool) Shutdown() {
	p.queue.Close()
	p.workers.Wait()
}

func (p *WorkerPool) worker() {
	defer p.workers.Done()

	for {
		task, ok := p.queue.Dequeue()
		if !ok {
			return
		}

		runTask(task)
	}
}

func runTask(task Task) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("worker task panic: %v\n%s", recovered, debug.Stack())
		}
	}()

	task()
}

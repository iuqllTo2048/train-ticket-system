package datastructure

import "sync"

type BlockingQueue[T any] struct {
	queue *RingQueue[T]

	mu sync.Mutex

	notEmpty *sync.Cond
	notFull  *sync.Cond

	closed bool
}

func NewBlockingQueue[T any](capacity int) *BlockingQueue[T] {
	q := &BlockingQueue[T]{
		queue: NewRingQueue[T](capacity),
	}

	q.notEmpty = sync.NewCond(&q.mu)
	q.notFull = sync.NewCond(&q.mu)

	return q
}

func (q *BlockingQueue[T]) Enqueue(value T) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 队列已满且尚未关闭，进入等待
	for q.queue.IsFull() && !q.closed {
		q.notFull.Wait()
	}

	// 队列已关闭，拒绝入队
	if q.closed {
		return false
	}

	// 调用原有环形队列的入队方法
	q.queue.Enqueue(value)

	// 唤醒一个等待任务的消费者
	q.notEmpty.Signal()

	return true
}

func (q *BlockingQueue[T]) Dequeue() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var zero T

	// 队列为空且未关闭，等待生产者入队
	for q.queue.IsEmpty() && !q.closed {
		q.notEmpty.Wait()
	}

	// 队列已经关闭，并且没有剩余元素
	if q.queue.IsEmpty() {
		return zero, false
	}

	// 调用原有环形队列的出队方法
	value, _ := q.queue.Dequeue()

	// 通知一个等待空位的生产者
	q.notFull.Signal()

	return value, true
}

func (q *BlockingQueue[T]) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 防止重复关闭
	if q.closed {
		return
	}

	q.closed = true

	// 唤醒所有等待中的消费者
	q.notEmpty.Broadcast()

	// 唤醒所有等待中的生产者
	q.notFull.Broadcast()
}

// Len 获取当前队列长度
func (q *BlockingQueue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.queue.Len()
}

// Capacity 获取队列容量
func (q *BlockingQueue[T]) Capacity() int {
	return q.queue.Capacity()
}

// IsClosed 判断队列是否关闭
func (q *BlockingQueue[T]) IsClosed() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.closed
}

// Snapshot 获取队列快照
func (q *BlockingQueue[T]) Snapshot() []T {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.queue.Snapshot()
}

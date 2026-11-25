package datastructure

// RingQueue 环形队列
// T 表示队列中存储的元素类型
type RingQueue[T any] struct {
	data     []T
	head     int
	tail     int
	size     int
	capacity int
}

// NewRingQueue 创建指定容量的环形队列
func NewRingQueue[T any](capacity int) *RingQueue[T] {
	if capacity <= 0 {
		panic("queue capacity must be greater than 0")
	}

	return &RingQueue[T]{
		data:     make([]T, capacity),
		head:     0,
		tail:     0,
		size:     0,
		capacity: capacity,
	}
}

// Enqueue 入队
func (q *RingQueue[T]) Enqueue(value T) bool {
	if q.IsFull() {
		return false
	}

	q.data[q.tail] = value

	q.tail = (q.tail + 1) % q.capacity

	q.size++

	return true
}

// Dequeue 出队
func (q *RingQueue[T]) Dequeue() (T, bool) {
	var zero T

	if q.IsEmpty() {
		return zero, false
	}

	value := q.data[q.head]

	q.data[q.head] = zero

	q.head = (q.head + 1) % q.capacity

	q.size--

	return value, true
}

// Peek 查看队首元素，但不删除
func (q *RingQueue[T]) Peek() (T, bool) {
	var zero T

	if q.IsEmpty() {
		return zero, false
	}

	return q.data[q.head], true
}

// IsEmpty 判断队列是否为空
func (q *RingQueue[T]) IsEmpty() bool {
	return q.size == 0
}

// IsFull 判断队列是否已满
func (q *RingQueue[T]) IsFull() bool {
	return q.size == q.capacity
}

// Len 获取当前有效元素数量
func (q *RingQueue[T]) Len() int {
	return q.size
}

// Capacity 获取队列最大容量
func (q *RingQueue[T]) Capacity() int {
	return q.capacity
}

// Snapshot 按照 FIFO 顺序返回队列元素的副本
func (q *RingQueue[T]) Snapshot() []T {
	result := make([]T, q.size)

	for i := 0; i < q.size; i++ {
		index := (q.head + i) % q.capacity

		result[i] = q.data[index]
	}

	return result
}

package datastructure

// Heap 泛型二叉堆。
type Heap[T any] struct {
	data []T
	less func(a, b T) bool
}

// NewHeap 创建一个空堆。
// less(a, b) 为 true 表示 a 的堆顶优先级高于 b。
func NewHeap[T any](less func(a, b T) bool) *Heap[T] {
	return &Heap[T]{
		data: make([]T, 0),
		less: less,
	}
}

// Len 返回堆中元素数量。
func (h *Heap[T]) Len() int {
	return len(h.data)
}

// Peek 查看堆顶元素，但不删除。
func (h *Heap[T]) Peek() (T, bool) {
	if len(h.data) == 0 {
		var zero T
		return zero, false
	}

	return h.data[0], true
}

// Push 插入一个元素。
func (h *Heap[T]) Push(value T) {
	h.data = append(h.data, value)

	h.siftUp(len(h.data) - 1)
}

// Pop 删除并返回堆顶元素。
func (h *Heap[T]) Pop() (T, bool) {
	if len(h.data) == 0 {
		var zero T
		return zero, false
	}

	result := h.data[0]

	last := len(h.data) - 1

	// 将最后一个元素移动到堆顶。
	h.data[0] = h.data[last]

	// 清除末尾元素，避免保留无用引用。
	var zero T
	h.data[last] = zero

	// 缩短 Slice。
	h.data = h.data[:last]

	// 如果堆不为空，执行下沉。
	if len(h.data) > 0 {
		h.siftDown(0)
	}

	return result, true
}

// siftUp 上浮操作。
func (h *Heap[T]) siftUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2

		// 当前元素不需要继续上浮。
		if !h.less(h.data[i], h.data[parent]) {
			break
		}

		// 交换当前节点和父节点。
		h.data[i], h.data[parent] =
			h.data[parent], h.data[i]

		// 更新当前节点位置。
		i = parent
	}
}

// siftDown 下沉操作。
func (h *Heap[T]) siftDown(i int) {
	n := len(h.data)

	for {
		left := 2*i + 1
		right := 2*i + 2

		best := i

		// 判断左儿子的优先级。
		if left < n &&
			h.less(h.data[left], h.data[best]) {
			best = left
		}

		// 判断右儿子的优先级。
		if right < n &&
			h.less(h.data[right], h.data[best]) {
			best = right
		}

		// 当前节点已经满足堆序性质。
		if best == i {
			break
		}

		// 与优先级最高的儿子交换。
		h.data[i], h.data[best] =
			h.data[best], h.data[i]

		// 继续向下调整。
		i = best
	}
}

package datastructure

import (
	"reflect"
	"testing"
)

func TestRingQueue(t *testing.T) {
	q := NewRingQueue[int](5)

	// 1. 验证初始状态
	if !q.IsEmpty() || q.Len() != 0 {
		t.Fatal("新队列应该为空")
	}

	// 2. 填满队列
	for _, value := range []int{10, 20, 30, 40, 50} {
		if !q.Enqueue(value) {
			t.Fatal("入队失败")
		}
	}

	if !q.IsFull() {
		t.Fatal("队列应该已满")
	}

	// 3. 验证队满时拒绝入队
	if q.Enqueue(999) {
		t.Fatal("已满队列不应该允许入队")
	}

	// 4. 连续出队两个元素
	for _, expected := range []int{10, 20} {
		value, ok := q.Dequeue()

		if !ok || value != expected {
			t.Fatalf("出队错误: %v", value)
		}
	}

	// 5. 重新利用数组前面的空闲位置
	if !q.Enqueue(60) || !q.Enqueue(70) {
		t.Fatal("环形入队失败")
	}

	// 6. 验证逻辑顺序
	expected := []int{30, 40, 50, 60, 70}

	actual := q.Snapshot()

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("快照错误: %v", actual)
	}

	t.Logf("环形队列物理存储: %v", q.data)
	t.Logf("环形队列逻辑顺序: %v", actual)
	t.Logf("head=%d tail=%d size=%d",
		q.head, q.tail, q.size)

	// 7. 验证 Peek 不删除元素
	value, ok := q.Peek()

	if !ok || value != 30 || q.Len() != 5 {
		t.Fatal("Peek 操作错误")
	}

	// 8. 验证 FIFO 顺序
	for _, expectedValue := range expected {
		value, ok := q.Dequeue()

		if !ok || value != expectedValue {
			t.Fatalf("FIFO 顺序错误: %v", value)
		}
	}

	// 9. 验证队空
	if !q.IsEmpty() {
		t.Fatal("队列应该为空")
	}

	if _, ok := q.Dequeue(); ok {
		t.Fatal("空队列不应该允许出队")
	}
}

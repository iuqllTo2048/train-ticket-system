package datastructure

import (
	"reflect"
	"testing"
)

func TestMinHeap(t *testing.T) {
	h := NewHeap(func(a, b int) bool {
		return a < b
	})

	for _, v := range []int{50, 20, 80, 10, 30} {
		h.Push(v)
	}

	result := make([]int, 0)

	for h.Len() > 0 {
		v, _ := h.Pop()
		result = append(result, v)
	}

	expected := []int{10, 20, 30, 50, 80}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("小顶堆错误: %v", result)
	}

	t.Logf("小顶堆取出顺序: %v", result)
}

func TestMaxHeap(t *testing.T) {
	h := NewHeap(func(a, b int) bool {
		return a > b
	})

	for _, v := range []int{50, 20, 80, 10, 30} {
		h.Push(v)
	}

	top, ok := h.Peek()

	if !ok || top != 80 || h.Len() != 5 {
		t.Fatal("Peek 操作错误")
	}

	result := make([]int, 0)

	for h.Len() > 0 {
		v, _ := h.Pop()
		result = append(result, v)
	}

	expected := []int{80, 50, 30, 20, 10}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("大顶堆错误: %v", result)
	}

	if _, ok := h.Pop(); ok {
		t.Fatal("空堆不应该能够 Pop")
	}

	t.Logf("大顶堆取出顺序: %v", result)
}

func TestHeapStruct(t *testing.T) {
	type Train struct {
		ID    string
		Price int
	}

	h := NewHeap(func(a, b Train) bool {
		return a.Price < b.Price
	})

	h.Push(Train{"G101", 88})
	h.Push(Train{"G102", 120})
	h.Push(Train{"G103", 65})

	top, ok := h.Peek()

	if !ok || top.ID != "G103" {
		t.Fatal("结构体比较错误")
	}

	t.Logf("最低票价车次: %+v", top)
}

package datastructure

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestBlockingQueueBasic(t *testing.T) {
	q := NewBlockingQueue[int](3)

	q.Enqueue(10)
	q.Enqueue(20)
	q.Enqueue(30)

	if q.Len() != 3 {
		t.Fatal("队列长度错误")
	}

	if q.Capacity() != 3 {
		t.Fatal("队列容量错误")
	}

	expected := []int{10, 20, 30}

	if !reflect.DeepEqual(q.Snapshot(), expected) {
		t.Fatal("队列快照错误")
	}

	for _, expectedValue := range expected {
		value, ok := q.Dequeue()

		if !ok || value != expectedValue {
			t.Fatalf("出队错误: %v", value)
		}
	}

	if q.Len() != 0 {
		t.Fatal("队列应该为空")
	}

	t.Log("BlockingQueue 基础功能测试通过")
}

func TestBlockingQueueConsumerWait(t *testing.T) {
	q := NewBlockingQueue[int](1)
	defer q.Close()

	result := make(chan int, 1)

	// 启动消费者协程
	go func() {
		value, ok := q.Dequeue()

		if ok {
			result <- value
		}
	}()

	// 验证消费者没有提前返回结果
	select {
	case value := <-result:
		t.Fatalf("消费者不应该提前返回: %d", value)

	case <-time.After(100 * time.Millisecond):
		t.Log("空队列：尚未收到消费者结果")
	}

	// 主协程作为生产者，执行入队
	q.Enqueue(100)

	// 验证消费者是否成功被唤醒
	select {
	case value := <-result:
		if value != 100 {
			t.Fatalf("出队数据错误: %d", value)
		}

		t.Logf("消费者被唤醒，成功出队: %d", value)

	case <-time.After(time.Second):
		t.Fatal("消费者未能及时唤醒")
	}
}

func TestBlockingQueueProducerWait(t *testing.T) {
	q := NewBlockingQueue[int](1)
	defer q.Close()

	q.Enqueue(100)

	result := make(chan bool, 1)

	// 启动生产者，尝试向满队列入队
	go func() {
		result <- q.Enqueue(200)
	}()

	// 验证生产者没有提前结束
	select {
	case <-result:
		t.Fatal("队列已满，生产者不应该提前返回")

	case <-time.After(100 * time.Millisecond):
		t.Log("满队列：生产者尚未返回")
	}

	// 主协程作为消费者，释放一个空位
	value, ok := q.Dequeue()

	if !ok || value != 100 {
		t.Fatal("第一次出队错误")
	}

	// 验证生产者是否恢复执行
	select {
	case success := <-result:
		if !success {
			t.Fatal("生产者入队失败")
		}

		t.Log("生产者被唤醒，成功入队 200")

	case <-time.After(2 * time.Second):
		t.Fatal("生产者未被唤醒")
	}

	value, ok = q.Dequeue()

	if !ok || value != 200 {
		t.Fatal("第二次出队错误")
	}
}

func TestBlockingQueueConcurrent(t *testing.T) {
	q := NewBlockingQueue[int](5)
	defer q.Close()

	const producerCount = 4
	const consumerCount = 3
	const tasksPerProducer = 25

	const total = producerCount * tasksPerProducer

	var producerWG sync.WaitGroup
	var consumerWG sync.WaitGroup

	results := make(chan int, total)

	// 1. 启动消费者
	for i := 0; i < consumerCount; i++ {
		consumerWG.Add(1)

		go func() {
			defer consumerWG.Done()

			for {
				value, ok := q.Dequeue()

				if !ok {
					return
				}

				results <- value
			}
		}()
	}

	// 2. 启动生产者
	for i := 0; i < producerCount; i++ {
		producerWG.Add(1)

		go func(id int) {
			defer producerWG.Done()

			for j := 0; j < tasksPerProducer; j++ {
				value := id*tasksPerProducer + j

				if !q.Enqueue(value) {
					return
				}
			}
		}(i)
	}

	// 3. 等待所有生产者完成
	producerWG.Wait()

	// 4. 关闭队列，不再接受新任务
	q.Close()

	// 5. 等待消费者处理完剩余任务并退出
	consumerWG.Wait()

	close(results)

	// 6. 统计每个元素出现的次数
	counts := make(map[int]int)

	count := 0

	for value := range results {
		counts[value]++
		count++
	}

	// 7. 检查元素总数
	if count != total {
		t.Fatalf("元素总数错误: %d", count)
	}

	// 8. 检查是否出现数据丢失或重复
	for i := 0; i < total; i++ {
		if counts[i] != 1 {
			t.Fatalf(
				"元素 %d 出现次数错误: %d",
				i,
				counts[i],
			)
		}
	}

	// 9. 检查队列是否已清空
	if q.Len() != 0 {
		t.Fatal("队列未清空")
	}

	t.Logf("生产者数量: %d", producerCount)
	t.Logf("消费者数量: %d", consumerCount)
	t.Logf("成功处理元素: %d", count)
	t.Log("无元素丢失，无重复元素")
}

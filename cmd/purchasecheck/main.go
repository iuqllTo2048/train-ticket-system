package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL      = "http://127.0.0.1:8080"
	requestCount = 100
	totalSeats   = 10
)

// 多个 Goroutine 可以共享同一个 HTTP Client。
var client = &http.Client{
	Timeout: 30 * time.Second,
}

// 保存单次请求的执行结果。
type PurchaseResult struct {
	StatusCode int
	Body       string
	Err        error
}

// 封装发送 POST JSON 请求的方法。
func postJSON(url string, data any) PurchaseResult {

	body, err := json.Marshal(data)
	if err != nil {
		return PurchaseResult{Err: err}
	}

	resp, err := client.Post(
		url,
		"application/json",
		bytes.NewReader(body),
	)

	if err != nil {
		return PurchaseResult{Err: err}
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)

	return PurchaseResult{
		StatusCode: resp.StatusCode,
		Body:       string(content),
		Err:        err,
	}
}

func main() {

	// 每次测试创建一个独立车次，避免污染之前的数据。
	trainID := fmt.Sprintf("TEST-%d", time.Now().UnixNano())

	// 第一步：创建只有 10 张票的测试车次。
	createResult := postJSON(
		baseURL+"/api/trains",
		map[string]any{
			"id":           trainID,
			"from_station": "成都东",
			"to_station":   "重庆北",
			"price_cents":  8850,
			"total_seats":  totalSeats,
		},
	)

	if createResult.Err != nil ||
		createResult.StatusCode != http.StatusCreated {

		log.Fatalf("创建车次失败: %+v", createResult)
	}

	fmt.Println("测试车次:", trainID)
	fmt.Println("初始余票:", totalSeats)
	fmt.Println("并发请求数量:", requestCount)

	// 第二步：准备并发测试。
	var wg sync.WaitGroup

	// 用于统一通知所有 Goroutine 开始抢票。
	start := make(chan struct{})

	// 每个 Goroutine 将结果写入自己的数组位置。
	results := make([]PurchaseResult, requestCount)

	purchaseURL := baseURL + "/api/trains/" +
		trainID + "/purchase"

	wg.Add(requestCount)

	// 第三步：创建 100 个 Goroutine。
	for i := 0; i < requestCount; i++ {

		go func(index int) {

			defer wg.Done()

			// 等待统一开始信号。
			<-start

			userID := fmt.Sprintf("U%03d", index+1)

			results[index] = postJSON(
				purchaseURL,
				map[string]any{
					"user_id":  userID,
					"quantity": 1,
				},
			)

		}(i)
	}

	// 第四步：开始并发抢票。
	begin := time.Now()

	close(start)

	// 等待所有购票请求执行结束。
	wg.Wait()

	cost := time.Since(begin)

	// 第五步：统计执行结果。
	success := 0
	insufficient := 0
	unexpected := 0

	for i, result := range results {

		switch {

		case result.Err != nil:

			unexpected++

			fmt.Printf(
				"用户 %d 请求异常: %v\n",
				i+1,
				result.Err,
			)

		case result.StatusCode == http.StatusCreated:

			success++

		case result.StatusCode == http.StatusConflict:

			insufficient++

		default:

			unexpected++

			fmt.Printf(
				"用户 %d 未预期响应: %d %s\n",
				i+1,
				result.StatusCode,
				result.Body,
			)
		}
	}

	fmt.Println("\n========== 测试结果 ==========")
	fmt.Println("测试车次:", trainID)
	fmt.Println("请求总数:", requestCount)
	fmt.Println("购票成功:", success)
	fmt.Println("库存冲突:", insufficient)
	fmt.Println("其他异常:", unexpected)
	fmt.Println("总耗时:", cost)

	// 第六步：判断 HTTP 层测试是否通过。
	if success == totalSeats &&
		insufficient == requestCount-totalSeats &&
		unexpected == 0 {

		fmt.Println("HTTP 并发测试通过，请继续验证数据库")
	} else {
		log.Fatal("HTTP 并发测试未达到预期，请检查原因")
	}
}

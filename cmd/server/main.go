package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/train-ticket-system/backend-app/internal/handler"
	"example.com/train-ticket-system/backend-app/internal/repository"
	"example.com/train-ticket-system/backend-app/internal/service"
	"example.com/train-ticket-system/backend-app/internal/taskmanager"
	"example.com/train-ticket-system/backend-app/internal/workerpool"

	"github.com/gin-gonic/gin"
	drivermysql "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	user := os.Getenv("MYSQL_USER")

	if user == "" {
		log.Fatal("请先设置 MYSQL_USER 环境变量")
	}

	// 构造 MySQL 连接配置。
	cfg := drivermysql.Config{
		User:   user,
		Passwd: os.Getenv("MYSQL_PASSWORD"),

		Net: "tcp",
		Addr: envOrDefault(
			"MYSQL_ADDR",
			"127.0.0.1:3306",
		),

		DBName: "train_ticket_app",

		ParseTime: true,
		Loc:       time.Local,

		Params: map[string]string{
			"charset": "utf8mb4",
		},
	}

	// 初始化 GORM。
	db, err := gorm.Open(
		gormmysql.Open(cfg.FormatDSN()),
		&gorm.Config{
			TranslateError: true,
		},
	)

	if err != nil {
		log.Fatalf("连接 MySQL 失败: %v", err)
	}

	// 获取底层 database/sql 连接池。
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库连接池失败: %v", err)
	}

	defer sqlDB.Close()

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 验证数据库连接。
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("数据库 Ping 失败: %v", err)
	}

	// 初始化异步任务调度组件。
	pool := workerpool.NewWorkerPool(3, 8)
	defer pool.Shutdown()

	taskManager := taskmanager.NewTaskManager(pool)
	taskHandler := handler.NewTaskHandler(taskManager)

	// 车次手动依赖注入。
	trainRepo := repository.NewGORMTrainRepository(db)
	trainSvc := service.NewTrainService(trainRepo)

	trainHandler := handler.NewTrainHandler(
		trainSvc,
		taskManager,
	)

	//  订单查询依赖注入
	orderRepo := repository.NewGORMOrderRepository(db)
	orderSvc := service.NewOrderService(orderRepo)
	orderHandler := handler.NewOrderHandler(orderSvc)

	// Gin 路由。
	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	trainHandler.RegisterRoutes(router)
	orderHandler.RegisterRoutes(router)
	taskHandler.RegisterRoutes(router)

	log.Println(
		"backend-app 已连接 MySQL，HTTP 地址 http://127.0.0.1:8080",
	)

	// 创建 HTTP Server。
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// 监听系统退出信号。
	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// 在独立协程中启动 HTTP Server。
	serverErr := make(chan error, 1)

	go func() {
		serverErr <- server.ListenAndServe()
	}()

	// 等待退出信号或者服务器运行错误。
	select {
	case <-signalCtx.Done():
		log.Println("收到退出信号，开始关闭服务")

	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP Server 运行错误: %v", err)
		}
	}

	// 停止接收新请求，等待正在处理的 HTTP 请求结束。
	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP Server 关闭失败: %v", err)
	}

	log.Println("HTTP Server 已关闭，准备释放后台资源")
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}

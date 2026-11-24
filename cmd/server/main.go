package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/train-ticket-system/backend-app/internal/handler"
	"example.com/train-ticket-system/backend-app/internal/repository"
	"example.com/train-ticket-system/backend-app/internal/service"

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

	// 手动依赖注入。
	repo := repository.NewGORMTrainRepository(db)
	svc := service.NewTrainService(repo)
	trainHandler := handler.NewTrainHandler(svc)

	// Gin 路由。
	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	trainHandler.RegisterRoutes(router)

	log.Println(
		"backend-app 已连接 MySQL，HTTP 地址 http://127.0.0.1:8080",
	)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}

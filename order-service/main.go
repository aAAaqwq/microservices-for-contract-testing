package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/model"
	"order-service/internal/repository"
	"order-service/internal/service"
)

// Hook 定义关闭时需要执行的函数
type Hook func(ctx context.Context) error

// GracefulShutdown 优雅关闭服务
func GracefulShutdown(ctx context.Context, timeout time.Duration, hooks ...Hook) {
	// 创建通知通道
	quit := make(chan os.Signal, 1)
	// 监听中断信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 执行所有关闭钩子
	for _, hook := range hooks {
		if err := hook(ctx); err != nil {
			log.Printf("Error during shutdown: %v\n", err)
		}
	}

	log.Println("Server exiting")
}

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("../default-config.yaml")
	if err != nil {
		log.Fatal("Failed to load order-db-config:", err)
	}

	// 连接数据库
	db, err := setupDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 初始化依赖
	repo := repository.NewOrderRepository(db)
	svc := service.NewOrderService(repo, cfg)
	handler := handler.NewOrderHandler(svc)

	// 设置路由
	r := setupRouter(handler)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// 在goroutine中启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 定义关闭钩子
	hooks := []Hook{
		// 关闭HTTP服务器
		func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
		// 关闭数据库连接
		func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	}

	// 优雅关闭
	GracefulShutdown(context.Background(), 10*time.Second, hooks...)
}

func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	// MySQL的DSN格式
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&autocommit=true",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})

	if err != nil {
		return nil, err
	}

	// 自动迁移
	err = db.AutoMigrate(&model.Order{}, &model.OrderItem{})
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %v", err)
	}

	return db, nil
}

func setupRouter(h *handler.OrderHandler) *gin.Engine {
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API路由
	api := r.Group("/api/v1")
	{
		api.POST("/orders", h.CreateOrder)
		api.GET("/orders/:id", h.GetOrder)
		api.GET("/orders/user/:userId", h.GetUserOrders)
		api.PUT("/orders/:id/status", h.UpdateOrderStatus)
		api.DELETE("/orders/:id", h.CancelOrder)
	}

	return r
}

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
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"notification-service/internal/cache"
	"notification-service/internal/config"
	"notification-service/internal/handler"
	"notification-service/internal/repository"
	"notification-service/internal/service"
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
		log.Fatal("Failed to load config:", err)
	}

	// 连接MongoDB
	mongoClient, err := setupMongoDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// 连接Redis
	redisClient := setupRedis(cfg)
	defer redisClient.Close()

	// 初始化依赖
	db := mongoClient.Database(cfg.MongoDB.Database)
	repo := repository.NewNotificationRepository(db)
	cacheClient := cache.NewNotificationCache(redisClient)
	svc := service.NewNotificationService(repo, cacheClient, cfg)
	handler := handler.NewNotificationHandler(svc)

	// 启动通知处理worker
	startNotificationWorker(svc)

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
		// 关闭worker
		func(ctx context.Context) error {
			svc.Stop() // 停止worker
			return nil
		},
		// 关闭MongoDB连接
		func(ctx context.Context) error {
			return mongoClient.Disconnect(ctx)
		},
		// 关闭Redis连接
		func(ctx context.Context) error {
			return redisClient.Close()
		},
	}

	// 优雅关闭
	GracefulShutdown(context.Background(), 10*time.Second, hooks...)
}

func setupMongoDB(cfg *config.Config) (*mongo.Client, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin",
		cfg.MongoDB.Username,
		cfg.MongoDB.Password,
		cfg.MongoDB.Host,
		cfg.MongoDB.Port,
		cfg.MongoDB.Database,
	)

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// 测试连接
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func setupRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}

func setupRouter(h *handler.NotificationHandler) *gin.Engine {
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API路由
	api := r.Group("/api/v1")
	{
		api.POST("/notifications", h.CreateNotification)
		api.GET("/notifications/:id", h.GetNotification)
		api.GET("/notifications/user/:userId", h.GetUserNotifications)
		api.POST("/notifications/batch", h.BatchCreateNotifications)
	}

	return r
}

func startNotificationWorker(svc *service.NotificationService) {
	go func() {
		ctx := context.Background()
		for {
			select {
			case <-svc.Done(): // 监听关闭信号
				log.Println("Notification worker shutting down...")
				return
			default:
				notification, err := svc.ProcessQueuedNotification(ctx)
				if err != nil {
					log.Printf("Error processing notification: %v", err)
					time.Sleep(time.Second * 5)
					continue
				}
				if notification == nil {
					time.Sleep(time.Second)
					continue
				}
			}
		}
	}()
}

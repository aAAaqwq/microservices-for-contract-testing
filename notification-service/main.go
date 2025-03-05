package main

import (
	"context"
	"fmt"
	"log"
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

	// 启动服务器
	r.Run(":" + cfg.Port)
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
	}()
}

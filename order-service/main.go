package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/model"
	"order-service/internal/repository"
	"order-service/internal/service"
)

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

	// 启动服务器
	r.Run(":" + cfg.Port)
}

func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	// MySQL的DSN格式
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
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

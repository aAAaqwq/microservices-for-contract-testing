package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"user-service/internal/config"
	"user-service/internal/handler"
	"user-service/internal/model"
	"user-service/internal/repository"
	"user-service/internal/service"
)

func main() {

	// 加载配置
	cfg, err := config.LoadConfig("../default-config.yaml")
	if err != nil {
		log.Fatal("Failed to load user-db-config:", err)
	}

	// 连接数据库
	db, err := setupDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 初始化依赖
	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo)
	handler := handler.NewUserHandler(svc)

	// 设置路由
	r := setupRouter(handler)

	// 启动服务器
	r.Run(":" + cfg.Port)
}

func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.User, cfg.DB.Password, cfg.DB.DBName, cfg.DB.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移
	db.AutoMigrate(&model.User{})

	return db, nil
}

func setupRouter(h *handler.UserHandler) *gin.Engine {
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API路由
	api := r.Group("/api/v1")
	{
		api.POST("/users", h.CreateUser)
		api.GET("/users/:id", h.GetUser)
		api.PUT("/users/:id", h.UpdateUser)
		api.DELETE("/users/:id", h.DeleteUser)
	}

	return r
}

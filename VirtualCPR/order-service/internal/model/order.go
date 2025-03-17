package model

import (
	"time"
)

// Order 订单模型
type Order struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	UserID      uint        `json:"user_id" binding:"required"`
	Items       []OrderItem `json:"items" gorm:"foreignKey:OrderID" binding:"required,dive"`
	TotalAmount float64     `json:"total_amount"`
	Status      string      `json:"status"` // pending, processing, completed, failed, cancelled
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// OrderItem 订单项模型
type OrderItem struct {
	ID       uint    `json:"id" gorm:"primaryKey"`
	OrderID  uint    `json:"order_id"`
	Name     string  `json:"name" binding:"required"`
	Price    float64 `json:"price" binding:"required"`
	Quantity int     `json:"quantity" binding:"required"`
}

// OrderStatusRequest 订单状态更新请求
type OrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=processing completed failed cancelled"`
}

package model

import (
	"time"
)

// Payment 支付模型
type Payment struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	OrderID     uint      `json:"order_id" binding:"required"`
	UserID      uint      `json:"user_id" binding:"required"`
	Amount      float64   `json:"amount" binding:"required"`
	Status      string    `json:"status"`                          // pending, success, failed, refunded
	PaymentType string    `json:"payment_type" binding:"required"` // credit_card, alipay, wechat
	TradeNo     string    `json:"trade_no,omitempty"`              // 第三方支付平台交易号
	PaidAt      time.Time `json:"paid_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PaymentRequest 支付请求
type PaymentRequest struct {
	OrderID     uint    `json:"order_id" binding:"required"`
	UserID      uint    `json:"user_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	PaymentType string  `json:"payment_type" binding:"required,oneof=credit_card alipay wechat"`
}

// RefundRequest 退款请求
type RefundRequest struct {
	Reason string `json:"reason" binding:"required"`
}

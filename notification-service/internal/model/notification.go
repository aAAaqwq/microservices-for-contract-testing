package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification 通知模型
type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    uint               `bson:"user_id" json:"user_id"`
	Type      string             `bson:"type" json:"type"` // email, sms
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	Status    string             `bson:"status" json:"status"`       // pending, sent, failed
	Recipient string             `bson:"recipient" json:"recipient"` // email address or phone number
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	SentAt    *time.Time         `bson:"sent_at,omitempty" json:"sent_at,omitempty"`
}

// NotificationRequest 通知请求
type NotificationRequest struct {
	UserID    uint   `json:"user_id" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=email sms"`
	Title     string `json:"title" binding:"required"`
	Content   string `json:"content" binding:"required"`
	Recipient string `json:"recipient" binding:"required"`
}

// BatchNotificationRequest 批量通知请求
type BatchNotificationRequest struct {
	Notifications []NotificationRequest `json:"notifications" binding:"required,dive"`
}

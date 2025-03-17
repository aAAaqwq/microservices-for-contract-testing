package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"notification-service/internal/model"

	"github.com/go-redis/redis/v8"
)

type NotificationCache struct {
	client *redis.Client
}

func NewNotificationCache(client *redis.Client) *NotificationCache {
	return &NotificationCache{client: client}
}

// CacheNotification 缓存通知
func (c *NotificationCache) CacheNotification(ctx context.Context, notification *model.Notification) error {
	key := fmt.Sprintf("notification:%s", notification.ID.Hex())
	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, 24*time.Hour).Err()
}

// GetCachedNotification 获取缓存的通知
func (c *NotificationCache) GetCachedNotification(ctx context.Context, id string) (*model.Notification, error) {
	key := fmt.Sprintf("notification:%s", id)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var notification model.Notification
	err = json.Unmarshal(data, &notification)
	if err != nil {
		return nil, err
	}

	return &notification, nil
}

// AddToQueue 添加到通知队列
func (c *NotificationCache) AddToQueue(ctx context.Context, notification *model.Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	return c.client.LPush(ctx, "notification_queue", data).Err()
}

// GetFromQueue 从队列获取通知
func (c *NotificationCache) GetFromQueue(ctx context.Context) (*model.Notification, error) {
	data, err := c.client.RPop(ctx, "notification_queue").Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var notification model.Notification
	err = json.Unmarshal(data, &notification)
	if err != nil {
		return nil, err
	}

	return &notification, nil
}

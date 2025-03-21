package service

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/gomail.v2"

	"notification-service/internal/cache"
	"notification-service/internal/config"
	"notification-service/internal/model"
	"notification-service/internal/repository"
)

type NotificationService struct {
	repo    *repository.NotificationRepository
	cache   *cache.NotificationCache
	config  *config.Config
	workers chan struct{} // 用于限制并发的worker数量
	done    chan struct{} // 添加用于控制worker关闭的通道
}

func NewNotificationService(repo *repository.NotificationRepository, cache *cache.NotificationCache, cfg *config.Config) *NotificationService {
	return &NotificationService{
		repo:    repo,
		cache:   cache,
		config:  cfg,
		workers: make(chan struct{}, 10), // 最多10个并发worker
		done:    make(chan struct{}),     // 初始化done通道
	}
}

// Done 返回用于监听服务关闭的通道
func (s *NotificationService) Done() <-chan struct{} {
	return s.done
}

// CreateNotification 创建通知
func (s *NotificationService) CreateNotification(ctx context.Context, req *model.NotificationRequest) (*model.Notification, error) {
	notification := &model.Notification{
		UserID:    req.UserID,
		Type:      req.Type,
		Title:     req.Title,
		Content:   req.Content,
		Status:    "pending",
		Recipient: req.Recipient,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, err
	}

	// 缓存通知
	if err := s.cache.CacheNotification(ctx, notification); err != nil {
		// 仅记录错误，不影响主流程
		fmt.Printf("Failed to cache notification: %v\n", err)
	}

	// 异步发送通知
	go s.processNotification(notification)

	return notification, nil
}

// GetNotification 获取通知
func (s *NotificationService) GetNotification(ctx context.Context, id string) (*model.Notification, error) {
	// 先从缓存获取
	notification, err := s.cache.GetCachedNotification(ctx, id)
	if err != nil {
		fmt.Printf("Failed to get notification from cache: %v\n", err)
	}
	if notification != nil {
		return notification, nil
	}

	// 缓存未命中，从数据库获取
	notification, err = s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	if err := s.cache.CacheNotification(ctx, notification); err != nil {
		fmt.Printf("Failed to cache notification: %v\n", err)
	}

	return notification, nil
}

// GetUserNotifications 获取用户通知
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uint) ([]model.Notification, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// BatchCreateNotifications 批量创建通知
func (s *NotificationService) BatchCreateNotifications(ctx context.Context, req *model.BatchNotificationRequest) error {
	for _, notificationReq := range req.Notifications {
		notification := &model.Notification{
			UserID:    notificationReq.UserID,
			Type:      notificationReq.Type,
			Title:     notificationReq.Title,
			Content:   notificationReq.Content,
			Status:    "pending",
			Recipient: notificationReq.Recipient,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.repo.Create(ctx, notification); err != nil {
			return err
		}

		// 添加到通知队列
		if err := s.cache.AddToQueue(ctx, notification); err != nil {
			fmt.Printf("Failed to add notification to queue: %v\n", err)
		}
	}

	return nil
}

// processNotification 处理通知发送
func (s *NotificationService) processNotification(notification *model.Notification) {
	// 获取worker令牌
	select {
	case s.workers <- struct{}{}: // 获取令牌
		defer func() { <-s.workers }() // 释放令牌
	case <-s.done: // 服务正在关闭
		return
	}

	var err error
	switch notification.Type {
	case "email":
		err = s.sendEmail(notification)
	case "sms":
		err = s.sendSMS(notification)
	default:
		err = fmt.Errorf("unsupported notification type: %s", notification.Type)
	}

	status := "sent"
	if err != nil {
		status = "failed"
		fmt.Printf("Failed to send notification: %v\n", err)
	}

	// 检查服务是否正在关闭
	select {
	case <-s.done:
		return
	default:
		ctx := context.Background()
		if err := s.repo.UpdateStatus(ctx, notification.ID.Hex(), status); err != nil {
			fmt.Printf("Failed to update notification status: %v\n", err)
		}
	}
}

// sendEmail 发送邮件
func (s *NotificationService) sendEmail(notification *model.Notification) error {
	d := gomail.NewDialer(
		s.config.Email.Host,
		s.config.Email.Port,
		s.config.Email.Username,
		s.config.Email.Password,
	)
	d.SSL = true

	m := gomail.NewMessage()
	m.SetHeader("From", s.config.Email.From)
	m.SetHeader("To", notification.Recipient)
	m.SetHeader("Subject", notification.Title)
	m.SetBody("text/plain", notification.Content)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// sendSMS 发送短信
func (s *NotificationService) sendSMS(notification *model.Notification) error {
	// TODO: 实现短信发送逻辑
	return fmt.Errorf("SMS sending not implemented")
}

// ProcessQueuedNotification 处理队列中的通知
func (s *NotificationService) ProcessQueuedNotification(ctx context.Context) (*model.Notification, error) {
	// 检查服务是否正在关闭
	select {
	case <-s.done:
		return nil, fmt.Errorf("service is shutting down")
	default:
	}

	// 从队列获取通知
	notification, err := s.cache.GetFromQueue(ctx)
	if err != nil {
		return nil, err
	}
	if notification == nil {
		return nil, nil
	}

	// 处理通知
	s.processNotification(notification)

	return notification, nil
}

// Stop 停止服务
func (s *NotificationService) Stop() {
	close(s.done)
}

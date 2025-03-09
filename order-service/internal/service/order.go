package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"order-service/internal/config"
	"order-service/internal/model"
	"order-service/internal/repository"
)

type OrderService struct {
	repo   *repository.OrderRepository
	config *config.Config
	client *http.Client
}

func NewOrderService(repo *repository.OrderRepository, cfg *config.Config) *OrderService {
	return &OrderService{
		repo:   repo,
		config: cfg,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(order *model.Order) error {
	// 验证用户
	if err := s.validateUser(order.UserID); err != nil {
		return fmt.Errorf("invalid user: %v", err)
	}

	// 计算总金额
	var total float64
	for _, item := range order.Items {
		total += item.Price * float64(item.Quantity)
	}
	order.TotalAmount = total
	order.Status = "pending"
	order.CreatedAt = time.Now()

	// 创建订单
	if err := s.repo.Create(order); err != nil {
		return fmt.Errorf("failed to create order: %v", err)
	}

	// 异步创建支付
	go s.createPayment(order)

	return nil
}

// GetOrder 获取订单
func (s *OrderService) GetOrder(id uint) (*model.Order, error) {
	return s.repo.GetByID(id)
}

// GetUserOrders 获取用户订单
func (s *OrderService) GetUserOrders(userID uint) ([]model.Order, error) {
	return s.repo.GetByUserID(userID)
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(id uint, status string) error {
	order, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 验证状态转换
	if !isValidStatusTransition(order.Status, status) {
		return errors.New("invalid status transition")
	}

	order.Status = status
	order.UpdatedAt = time.Now()

	if err := s.repo.Update(order); err != nil {
		return err
	}

	// 发送通知
	go s.sendNotification(order)

	return nil
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(id uint) error {
	order, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if order.Status != "pending" {
		return errors.New("only pending orders can be cancelled")
	}

	order.Status = "cancelled"
	order.UpdatedAt = time.Now()

	if err := s.repo.Update(order); err != nil {
		return err
	}

	// 发送通知
	go s.sendNotification(order)

	return nil
}

// validateUser 验证用户,HTTP调用user-service
func (s *OrderService) validateUser(userID uint) error {
	resp, err := s.client.Get(fmt.Sprintf("%s/api/v1/users/%d", s.config.Services.UserServiceURL, userID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("user not found")
	}

	return nil
}

// createPayment 创建支付
func (s *OrderService) createPayment(order *model.Order) {
	paymentReq := map[string]interface{}{
		"order_id": order.ID,
		"user_id":  order.UserID,
		"amount":   order.TotalAmount,
		"payment_type": "wechat",//默认模拟wechat支付
	}

	jsonData, err := json.Marshal(paymentReq)
	if err != nil {
		return
	}
	// fmt.Println(string(jsonData))

	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/payments", s.config.Services.PaymentServiceURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// sendNotification 发送通知
func (s *OrderService) sendNotification(order *model.Order) {
	notificationReq := map[string]interface{}{
		"user_id": order.UserID,
		"type":    "email",
		"title":   fmt.Sprintf("Order %d Status Update", order.ID),
		"content": fmt.Sprintf("Your order #%d status has been updated to %s", order.ID, order.Status),
	}

	jsonData, err := json.Marshal(notificationReq)
	if err != nil {
		return
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/notifications", s.config.Services.NotificationServiceURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// isValidStatusTransition 验证状态转换是否有效
func isValidStatusTransition(from, to string) bool {
	transitions := map[string][]string{
		"pending":    {"processing", "cancelled"},
		"processing": {"completed", "failed"},
		"completed":  {},
		"failed":     {},
		"cancelled":  {},
	}

	validTransitions, exists := transitions[from]
	if !exists {
		return false
	}

	for _, validStatus := range validTransitions {
		if validStatus == to {
			return true
		}
	}

	return false
}

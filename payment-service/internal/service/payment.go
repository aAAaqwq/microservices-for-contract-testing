package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	// "sync"
	"time"

	"payment-service/internal/config"
	"payment-service/internal/model"
	"payment-service/internal/repository"

	"github.com/spf13/cast"
)

type PaymentService struct {
	repo   *repository.PaymentRepository
	config *config.Config
	client *http.Client
	// wg     sync.WaitGroup
}

func NewPaymentService(repo *repository.PaymentRepository, cfg *config.Config) *PaymentService {
	return &PaymentService{
		repo:   repo,
		config: cfg,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// CreatePayment 创建支付
func (s *PaymentService) CreatePayment(req *model.PaymentRequest) (*model.Payment, error) {
	// 验证订单
	if err := s.validateOrder(req.OrderID); err != nil {
		return nil, fmt.Errorf("invalid order: %v", err)
	}

	payment := &model.Payment{
		OrderID:     req.OrderID,
		UserID:      req.UserID,
		Amount:      req.Amount,
		PaymentType: req.PaymentType,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(payment); err != nil {
		return nil, err
	}
	// fmt.Println("处理支付前")

	// 异步处理支付
	go s.processPayment(payment)

	return payment, nil
}

// GetPayment 获取支付信息
func (s *PaymentService) GetPayment(id uint) (*model.Payment, error) {
	return s.repo.GetByID(id)
}

// GetOrderPayment 获取订单支付信息
func (s *PaymentService) GetOrderPayment(orderID uint) (*model.Payment, error) {
	return s.repo.GetByOrderID(orderID)
}

// GetUserPayments 获取用户支付记录
func (s *PaymentService) GetUserPayments(userID uint) ([]model.Payment, error) {
	return s.repo.GetByUserID(userID)
}

// ProcessRefund 处理退款
func (s *PaymentService) ProcessRefund(id uint, reason string) error {
	payment, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if payment.Status != "success" {
		return errors.New("only successful payments can be refunded")
	}

	// 处理退款逻辑
	if err := s.handleRefund(payment); err != nil {
		return err
	}

	payment.Status = "refunded"
	payment.UpdatedAt = time.Now()

	if err := s.repo.Update(payment); err != nil {
		return err
	}

	// 发送通知
	go s.sendNotification(payment, "refund")

	return nil
}

// validateOrder 验证订单
func (s *PaymentService) validateOrder(orderID uint) error {
	resp, err := s.client.Get(fmt.Sprintf("%s/api/v1/orders/%d", s.config.Services.OrderServiceURL, orderID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("order not found")
	}

	return nil
}

// processPayment 处理支付
func (s *PaymentService) processPayment(payment *model.Payment) {
	// fmt.Println("处理支付中")
	// 模拟正在支付处理
	time.Sleep(2 * time.Second)

	err := s.UpdateOrderStatus(payment.OrderID, "processing")
	if err != nil {
		log.Printf("%v",err)
		payment.Status = "failed"
		payment.UpdatedAt = time.Now()
		s.repo.Update(payment)
		return
	}

	// 模拟支付完成
	time.Sleep(2 * time.Second)

	err = s.UpdateOrderStatus(payment.OrderID, "completed")
	if err != nil {
		log.Printf("%v",err)
		payment.UpdatedAt = time.Now()
		payment.Status = "failed"
		s.repo.Update(payment)
		return
	}

	// 生成交易号
	payment.TradeNo = fmt.Sprintf("T%d%d", payment.ID, time.Now().Unix())
	payment.Status = "success"
	payment.PaidAt = time.Now()
	payment.UpdatedAt = time.Now()

	if err := s.repo.Update(payment); err != nil {
		return
	}
	// 发送通知
	go s.sendNotification(payment, "payment")
}

// UpdateOrderStatus 更新订单状态
func (s *PaymentService) UpdateOrderStatus(orderID uint, status string) error {
	jsonData, err := json.Marshal(map[string]string{"status": status})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/orders/%d/status", s.config.Services.OrderServiceURL, orderID), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to update order status")
	}

	return nil
}

// handleRefund 处理退款
func (s *PaymentService) handleRefund(payment *model.Payment) error {
	// 模拟退款处理
	time.Sleep(2 * time.Second)
	// 取消订单
	go s.UpdateOrderStatus(payment.OrderID, "cancelled")
	return nil
}

// sendNotification 发送通知
func (s *PaymentService) sendNotification(payment *model.Payment, notifyType string) {
	var title, content string
	switch notifyType {
	case "payment":
		title = fmt.Sprintf("Payment Success for Order #%d", payment.OrderID)
		content = fmt.Sprintf("Your payment of %.2f has been processed successfully", payment.Amount)
	case "refund":
		title = fmt.Sprintf("Refund Processed for Order #%d", payment.OrderID)
		content = fmt.Sprintf("Your refund of %.2f has been processed", payment.Amount)
	}
	email:=s.GetUserEmail(payment.UserID)

	notificationReq := map[string]interface{}{
		"user_id": payment.UserID,
		"type":    "email",
		"title":   title,
		"content": content,
		"recipient": email,
	}

	jsonData, err := json.Marshal(notificationReq)
	if err != nil {
		return
	}
	// fmt.Println(string(jsonData))

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

// GetUserEmail 获取用户邮箱
func (s *PaymentService) GetUserEmail(userID uint) string {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/users/%d", s.config.Services.UserServiceURL, userID))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	userInfo := map[string]interface{}{
		"email": "",
		"username": "",
		"id": 0,
	}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return ""
	}
	// fmt.Println(userInfo)
	return cast.ToString(userInfo["email"])
}

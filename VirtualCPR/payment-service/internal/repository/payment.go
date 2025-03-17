package repository

import (
	"payment-service/internal/model"

	"gorm.io/gorm"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

func (r *PaymentRepository) GetByID(id uint) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) GetByOrderID(orderID uint) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.Where("order_id = ?", orderID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) GetByUserID(userID uint) ([]model.Payment, error) {
	var payments []model.Payment
	err := r.db.Where("user_id = ?", userID).Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *PaymentRepository) Update(payment *model.Payment) error {
	return r.db.Save(payment).Error
}

func (r *PaymentRepository) GetPendingPayments() ([]model.Payment, error) {
	var payments []model.Payment
	err := r.db.Where("status = ?", "pending").Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}

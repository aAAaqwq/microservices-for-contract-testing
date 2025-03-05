package repository

import (
	"order-service/internal/model"

	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Items").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) GetByUserID(userID uint) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Preload("Items").Where("user_id = ?", userID).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

func (r *OrderRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除订单项
		if err := tx.Where("order_id = ?", id).Delete(&model.OrderItem{}).Error; err != nil {
			return err
		}
		// 删除订单
		if err := tx.Delete(&model.Order{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

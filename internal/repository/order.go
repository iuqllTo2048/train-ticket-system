package repository

import (
	"context"
	"errors"

	"example.com/train-ticket-system/backend-app/internal/model"
	"gorm.io/gorm"
)

var ErrOrderNotFound = errors.New("订单不存在")

// OrderRepository 定义订单查询接口。
type OrderRepository interface {
	FindByOrderNo(
		ctx context.Context,
		orderNo string,
	) (model.Order, error)

	ListByUserID(
		ctx context.Context,
		userID string,
	) ([]model.Order, error)
}

// GORMOrderRepository 实现订单查询接口。
type GORMOrderRepository struct {
	db *gorm.DB
}

func NewGORMOrderRepository(
	db *gorm.DB,
) *GORMOrderRepository {
	return &GORMOrderRepository{db: db}
}

// FindByOrderNo 根据订单编号查询订单。
func (r *GORMOrderRepository) FindByOrderNo(
	ctx context.Context,
	orderNo string,
) (model.Order, error) {

	var order model.Order

	err := r.db.WithContext(ctx).
		Where("order_no = ?", orderNo).
		Take(&order).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Order{}, ErrOrderNotFound
	}

	return order, err
}

// ListByUserID 查询指定用户的全部订单。
func (r *GORMOrderRepository) ListByUserID(
	ctx context.Context,
	userID string,
) ([]model.Order, error) {

	orders := make([]model.Order, 0)

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Find(&orders).Error

	return orders, err
}

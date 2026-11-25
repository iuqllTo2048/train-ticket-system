package service

import (
	"context"
	"fmt"
	"strings"

	"example.com/train-ticket-system/backend-app/internal/model"
	"example.com/train-ticket-system/backend-app/internal/repository"
)

type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(
	repo repository.OrderRepository,
) *OrderService {
	return &OrderService{repo: repo}
}

// Get 查询单个订单。
func (s *OrderService) Get(
	ctx context.Context,
	orderNo string,
) (model.Order, error) {

	orderNo = strings.TrimSpace(orderNo)

	if orderNo == "" || len(orderNo) > 64 {
		return model.Order{}, fmt.Errorf(
			"%w: 订单编号不合法",
			ErrInvalidInput,
		)
	}

	return s.repo.FindByOrderNo(ctx, orderNo)
}

// ListByUser 查询用户的全部订单。
func (s *OrderService) ListByUser(
	ctx context.Context,
	userID string,
) ([]model.Order, error) {

	userID = strings.TrimSpace(userID)

	if userID == "" || len(userID) > 64 {
		return nil, fmt.Errorf(
			"%w: 用户编号不合法",
			ErrInvalidInput,
		)
	}

	return s.repo.ListByUserID(ctx, userID)
}

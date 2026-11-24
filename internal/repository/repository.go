package repository

import (
	"context"
	"errors"

	"example.com/train-ticket-system/backend-app/internal/model"
)

var (
	ErrNotFound = errors.New("车次不存在")

	ErrInsufficientSeats = errors.New("余票不足")

	ErrConflict = errors.New("数据冲突或存在购票记录")
)

type TrainRepository interface {

	// 查询单个车次
	FindByID(
		ctx context.Context,
		id string,
	) (model.Train, error)

	// 查询车次列表
	List(
		ctx context.Context,
		from, to string,
	) ([]model.Train, error)

	// 新增车次
	Create(
		ctx context.Context,
		train model.Train,
	) error

	// 更新车次基本信息
	UpdateInfo(
		ctx context.Context,
		id string,
		changes map[string]any,
	) error

	// 删除车次
	Delete(
		ctx context.Context,
		id string,
	) error

	// 原子扣票并生成购票流水
	Purchase(
		ctx context.Context,
		id string,
		quantity int,
	) (model.PurchaseLog, error)
}

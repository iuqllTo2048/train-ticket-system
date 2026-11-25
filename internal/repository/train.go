package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math"

	"example.com/train-ticket-system/backend-app/internal/model"
	"gorm.io/gorm"
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

	// 原子扣票,创建订单及购票流水
	Purchase(
		ctx context.Context,
		id string,
		userID string,
		quantity int,
	) (model.Order, error)
}

type GORMTrainRepository struct {
	db *gorm.DB
}

func NewGORMTrainRepository(db *gorm.DB) *GORMTrainRepository {
	return &GORMTrainRepository{db: db}
}

// FindByID 查询单个车次。
func (r *GORMTrainRepository) FindByID(
	ctx context.Context,
	id string,
) (model.Train, error) {

	var train model.Train

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Take(&train).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return train, ErrNotFound
	}

	return train, err
}

// List 查询车次列表。
func (r *GORMTrainRepository) List(
	ctx context.Context,
	from, to string,
) ([]model.Train, error) {

	trains := make([]model.Train, 0)

	query := r.db.WithContext(ctx).
		Model(&model.Train{})

	if from != "" {
		query = query.Where("from_station = ?", from)
	}

	if to != "" {
		query = query.Where("to_station = ?", to)
	}

	err := query.Order("id ASC").
		Find(&trains).Error

	return trains, err
}

// Create 插入新车次。
func (r *GORMTrainRepository) Create(
	ctx context.Context,
	train model.Train,
) error {

	err := r.db.WithContext(ctx).
		Create(&train).Error

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrConflict
	}

	return err
}

// UpdateInfo 更新车次的基本信息，不修改余票。
func (r *GORMTrainRepository) UpdateInfo(
	ctx context.Context,
	id string,
	changes map[string]any,
) error {

	result := r.db.WithContext(ctx).
		Model(&model.Train{}).
		Where("id = ?", id).
		Updates(changes)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		return nil
	}

	// MySQL 对“修改为原值”可能返回 0。
	// 此时需要进一步判断车次是否存在。
	_, err := r.FindByID(ctx, id)
	return err
}

// Delete 删除车次。
// 存在购票流水时，由外键约束禁止删除。
func (r *GORMTrainRepository) Delete(
	ctx context.Context,
	id string,
) error {

	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.Train{})

	if errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
		return ErrConflict
	}

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// generateOrderNo 生成随机订单编号。
func generateOrderNo() (string, error) {

	var data [16]byte

	_, err := rand.Read(data[:])

	if err != nil {
		return "", err
	}

	return "ORD-" + hex.EncodeToString(data[:]), nil
}

// Purchase 原子扣票，创建订单及购票流水
func (r *GORMTrainRepository) Purchase(
	ctx context.Context,
	id string,
	userID string,
	quantity int,
) (model.Order, error) {

	// 防御性校验，避免非法数量进入数据库操作。
	if quantity <= 0 || quantity > 10 {
		return model.Order{}, errors.New("非法购票数量")
	}

	var order model.Order

	// 在开启事务之前生成订单编号。
	orderNo, err := generateOrderNo()

	if err != nil {
		return model.Order{}, err
	}

	err = r.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {

			// 第一步：数据库原子条件更新。
			result := tx.Model(&model.Train{}).
				Where(
					"id = ? AND available_seats >= ?",
					id,
					quantity,
				).
				Update(
					"available_seats",
					gorm.Expr(
						"available_seats - ?",
						quantity,
					),
				)

			if result.Error != nil {
				return result.Error
			}

			// 第二步：判断是否成功扣票。
			if result.RowsAffected == 0 {

				var train model.Train

				err := tx.Where("id = ?", id).
					Take(&train).Error

				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrNotFound
				}

				if err != nil {
					return err
				}

				return ErrInsufficientSeats
			}

			// 第三步：在同一事务中查询车次信息。
			var train model.Train

			if err := tx.Where("id = ?", id).
				Take(&train).Error; err != nil {

				return err
			}

			// 防止计算订单金额时发生 int64 溢出。
			if train.PriceCents <= 0 ||
				train.PriceCents > math.MaxInt64/int64(quantity) {

				return errors.New("订单金额不合法")
			}

			// 第四步：构造正式订单对象。
			order = model.Order{
				OrderNo:          orderNo,
				TrainID:          id,
				UserID:           userID,
				Quantity:         quantity,
				UnitPriceCents:   train.PriceCents,
				TotalAmountCents: train.PriceCents * int64(quantity),
				Status:           model.OrderStatusSuccess,
			}

			// 第五步：写入正式订单。
			if err := tx.Create(&order).Error; err != nil {
				return err
			}

			// 第六步：保留原有购票流水。
			log := model.PurchaseLog{
				TrainID:  id,
				Quantity: quantity,
			}

			return tx.Create(&log).Error
		},
	)

	// 事务执行失败，返回空订单及错误。
	if err != nil {
		return model.Order{}, err
	}

	// 只有事务成功提交才返回正式订单。
	return order, nil
}

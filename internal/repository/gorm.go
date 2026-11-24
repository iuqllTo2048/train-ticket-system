package repository

import (
	"context"
	"errors"

	"example.com/train-ticket-system/backend-app/internal/model"
	"gorm.io/gorm"
)

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

// Purchase 扣减真实余票，并写入购票流水。
func (r *GORMTrainRepository) Purchase(
	ctx context.Context,
	id string,
	quantity int,
) (model.PurchaseLog, error) {

	var log model.PurchaseLog

	err := r.db.WithContext(ctx).Transaction(
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

			// 更新失败：进一步区分车次不存在与余票不足。
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

			// 第二步：写入购票流水。
			log = model.PurchaseLog{
				TrainID:  id,
				Quantity: quantity,
			}

			// 必须使用 tx，而不是 r.db。
			return tx.Create(&log).Error
		},
	)

	if err != nil {
		return model.PurchaseLog{}, err
	}

	return log, nil
}

package service

import (
	"context"
	"errors"
	"example.com/train-ticket-system/backend-app/internal/datastructure"
	"fmt"
	"strings"
	"sync"

	"example.com/train-ticket-system/backend-app/internal/model"
	"example.com/train-ticket-system/backend-app/internal/repository"
)

var ErrInvalidInput = errors.New("请求参数不合法")

type TrainService struct {
	repo repository.TrainRepository

	stationMu    sync.RWMutex
	stationIndex *datastructure.Trie
}

func NewTrainService(
	repo repository.TrainRepository,
) *TrainService {
	return &TrainService{repo: repo}
}

type CreateTrainInput struct {
	ID          string
	FromStation string
	ToStation   string
	PriceCents  int64
	TotalSeats  int
}

type TrainPatch struct {
	FromStation *string
	ToStation   *string
	PriceCents  *int64
}

// Get 查询车次。
func (s *TrainService) Get(
	ctx context.Context,
	id string,
) (model.Train, error) {
	return s.repo.FindByID(ctx, id)
}

// List 查询车次列表。
func (s *TrainService) List(
	ctx context.Context,
	from, to string,
) ([]model.Train, error) {

	return s.repo.List(
		ctx,
		strings.TrimSpace(from),
		strings.TrimSpace(to),
	)
}

// Create 新增车次。
func (s *TrainService) Create(
	ctx context.Context,
	in CreateTrainInput,
) (model.Train, error) {

	in.ID = strings.TrimSpace(in.ID)
	in.FromStation = strings.TrimSpace(in.FromStation)
	in.ToStation = strings.TrimSpace(in.ToStation)

	if len(in.ID) == 0 ||
		len(in.ID) > 32 ||
		len(in.FromStation) == 0 ||
		len(in.FromStation) > 64 ||
		len(in.ToStation) == 0 ||
		len(in.ToStation) > 64 ||
		in.FromStation == in.ToStation ||
		in.PriceCents <= 0 ||
		in.TotalSeats <= 0 {

		return model.Train{}, fmt.Errorf(
			"%w: 车次、站名、票价或座位数错误",
			ErrInvalidInput,
		)
	}

	train := model.Train{
		ID:             in.ID,
		FromStation:    in.FromStation,
		ToStation:      in.ToStation,
		PriceCents:     in.PriceCents,
		TotalSeats:     in.TotalSeats,
		AvailableSeats: in.TotalSeats,
	}

	s.stationMu.Lock()
	defer s.stationMu.Unlock()

	// 使旧索引失效。
	s.stationIndex = nil

	// 执行原有数据库写入操作。
	if err := s.repo.Create(ctx, train); err != nil {
		return model.Train{}, err
	}

	// 数据库写入成功后，重建站点索引。
	s.refreshStationIndexLocked(ctx)

	return s.repo.FindByID(ctx, train.ID)
}

// Update 更新车次信息。
// 不允许通过普通接口直接修改余票。
func (s *TrainService) Update(
	ctx context.Context,
	id string,
	patch TrainPatch,
) (model.Train, error) {

	if patch.FromStation == nil &&
		patch.ToStation == nil &&
		patch.PriceCents == nil {

		return model.Train{}, fmt.Errorf(
			"%w: 至少提供一个修改字段",
			ErrInvalidInput,
		)
	}

	s.stationMu.Lock()
	defer s.stationMu.Unlock()

	train, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return model.Train{}, err
	}

	changes := map[string]any{}

	if patch.FromStation != nil {
		train.FromStation = strings.TrimSpace(*patch.FromStation)
		changes["from_station"] = train.FromStation
	}

	if patch.ToStation != nil {
		train.ToStation = strings.TrimSpace(*patch.ToStation)
		changes["to_station"] = train.ToStation
	}

	if patch.PriceCents != nil {
		train.PriceCents = *patch.PriceCents
		changes["price_cents"] = train.PriceCents
	}

	if train.FromStation == "" ||
		train.ToStation == "" ||
		len(train.FromStation) > 64 ||
		len(train.ToStation) > 64 ||
		train.FromStation == train.ToStation ||
		train.PriceCents <= 0 {

		return model.Train{}, fmt.Errorf(
			"%w: 修改后的车次信息不合法",
			ErrInvalidInput,
		)
	}

	s.stationIndex = nil

	if err := s.repo.UpdateInfo(ctx, id, changes); err != nil {
		return model.Train{}, err
	}

	s.refreshStationIndexLocked(ctx)

	return s.repo.FindByID(ctx, id)
}

// Delete 删除车次。
func (s *TrainService) Delete(
	ctx context.Context,
	id string,
) error {
	s.stationMu.Lock()
	defer s.stationMu.Unlock()

	s.stationIndex = nil

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.refreshStationIndexLocked(ctx)

	return nil
}

// Purchase 执行购票业务，并返回正式订单。
func (s *TrainService) Purchase(
	ctx context.Context,
	id string,
	userID string,
	quantity int,
) (model.Order, error) {

	userID = strings.TrimSpace(userID)

	if userID == "" || len(userID) > 64 {
		return model.Order{}, fmt.Errorf(
			"%w: 用户编号不合法",
			ErrInvalidInput,
		)
	}

	if quantity <= 0 || quantity > 10 {
		return model.Order{}, fmt.Errorf(
			"%w: 每次购买数量必须在 1 到 10 之间",
			ErrInvalidInput,
		)
	}

	return s.repo.Purchase(
		ctx,
		id,
		userID,
		quantity,
	)
}

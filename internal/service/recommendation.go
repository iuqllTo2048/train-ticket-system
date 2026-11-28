package service

import (
	"context"
	"fmt"

	"example.com/train-ticket-system/backend-app/internal/datastructure"
	"example.com/train-ticket-system/backend-app/internal/model"
)

// RecommendTrains 推荐符合条件的 Top-K 车次。
func (s *TrainService) RecommendTrains(
	ctx context.Context,
	from string,
	to string,
	strategy string,
	k int,
) ([]model.Train, error) {

	// 1. 参数校验。
	if k <= 0 || k > 50 {
		return nil, fmt.Errorf(
			"%w: k 必须在 1 到 50 之间",
			ErrInvalidInput,
		)
	}

	// 2. 根据策略定义推荐优先级。
	var better func(a, b model.Train) bool

	switch strategy {
	case "price":
		better = func(a, b model.Train) bool {
			if a.PriceCents == b.PriceCents {
				return a.ID < b.ID
			}
			return a.PriceCents < b.PriceCents
		}

	case "seats":
		better = func(a, b model.Train) bool {
			if a.AvailableSeats == b.AvailableSeats {
				return a.ID < b.ID
			}
			return a.AvailableSeats > b.AvailableSeats
		}

	default:
		return nil, fmt.Errorf(
			"%w: 不支持的推荐策略",
			ErrInvalidInput,
		)
	}

	// 3. 从数据库查询候选车次。
	trains, err := s.List(ctx, from, to)

	if err != nil {
		return nil, err
	}

	// 4. 创建堆，堆顶保存最应该淘汰的车次。
	h := datastructure.NewHeap(
		func(a, b model.Train) bool {
			return better(b, a)
		},
	)

	// 5. 遍历候选车次，筛选 Top-K。
	for _, train := range trains {

		// 跳过已经售罄的车次。
		if train.AvailableSeats <= 0 {
			continue
		}

		// 堆未满，直接插入。
		if h.Len() < k {
			h.Push(train)
			continue
		}

		// 堆已满，获取当前最差的候选车次。
		worst, _ := h.Peek()

		// 新车次更值得推荐，则替换堆顶。
		if better(train, worst) {
			h.Pop()
			h.Push(train)
		}
	}

	// 6. 按照推荐顺序输出结果。
	result := make([]model.Train, h.Len())

	for i := len(result) - 1; i >= 0; i-- {
		train, _ := h.Pop()
		result[i] = train
	}

	return result, nil
}

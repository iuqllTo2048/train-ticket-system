package service

import (
	"context"
	"log"
	"strings"

	"example.com/train-ticket-system/backend-app/internal/datastructure"
)

// InitStationIndex 初始化站点索引。
func (s *TrainService) InitStationIndex(
	ctx context.Context,
) error {
	s.stationMu.Lock()
	defer s.stationMu.Unlock()

	return s.rebuildStationIndexLocked(ctx)
}

// rebuildStationIndexLocked 从数据库重建字典树。
// 调用者必须已经持有 stationMu 写锁。
func (s *TrainService) rebuildStationIndexLocked(
	ctx context.Context,
) error {
	trains, err := s.repo.List(ctx, "", "")

	if err != nil {
		return err
	}

	index := datastructure.NewTrie()

	for _, train := range trains {
		index.Insert(train.FromStation)
		index.Insert(train.ToStation)
	}

	s.stationIndex = index

	return nil
}

// SuggestStations 根据前缀推荐站点。
func (s *TrainService) SuggestStations(
	ctx context.Context,
	prefix string,
	limit int,
) ([]string, error) {
	prefix = strings.TrimSpace(prefix)

	// 优先使用已经存在的索引。
	s.stationMu.RLock()

	if s.stationIndex != nil {
		result := s.stationIndex.SearchPrefix(prefix, limit)

		s.stationMu.RUnlock()

		return result, nil
	}

	s.stationMu.RUnlock()

	// 索引不存在时，尝试重新构建。
	s.stationMu.Lock()
	defer s.stationMu.Unlock()

	if s.stationIndex == nil {
		if err := s.rebuildStationIndexLocked(ctx); err != nil {
			return nil, err
		}
	}

	return s.stationIndex.SearchPrefix(prefix, limit), nil
}

// refreshStationIndexLocked 在车次变更后刷新索引。
// 调用者必须已经持有 stationMu 写锁。
func (s *TrainService) refreshStationIndexLocked(
	ctx context.Context,
) {
	if err := s.rebuildStationIndexLocked(ctx); err != nil {
		s.stationIndex = nil

		log.Printf("站点索引刷新失败: %v", err)
	}
}

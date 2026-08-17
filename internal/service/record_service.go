package service

import (
	"sort"

	"approvalflow/internal/model"
)

// GetRecord 查询审批记录详情。
func (s *Service) GetRecord(id string) (*model.ApprovalRecord, error) {
	return s.store.GetRecord(id)
}

// ListRecords 分页查询审批记录，按时间倒序。
func (s *Service) ListRecords(filter model.RecordFilter, page, size int) ([]*model.ApprovalRecord, int, error) {
	all := s.store.ListRecords()
	matched := make([]*model.ApprovalRecord, 0, len(all))
	for _, r := range all {
		if filter.Match(r) {
			matched = append(matched, r)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.ApprovalRecord{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// Timeline 查询某审批单的完整审批时间线（按时间升序）。
func (s *Service) Timeline(requestID string) ([]*model.ApprovalRecord, error) {
	if _, err := s.store.GetRequest(requestID); err != nil {
		return nil, err
	}
	list := make([]*model.ApprovalRecord, 0)
	for _, r := range s.store.ListRecords() {
		if r.RequestID == requestID {
			list = append(list, r)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.Before(list[j].CreatedAt) })
	return list, nil
}

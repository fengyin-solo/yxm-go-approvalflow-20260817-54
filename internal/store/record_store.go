package store

import (
	"approvalflow/internal/model"
)

func (s *MemoryStore) CreateRecord(r *model.ApprovalRecord) error {
	s.records[r.ID] = r
	return nil
}

func (s *MemoryStore) GetRecord(id string) (*model.ApprovalRecord, error) {
	r, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return r, nil
}

func (s *MemoryStore) ListRecords() []*model.ApprovalRecord {
	list := make([]*model.ApprovalRecord, 0, len(s.records))
	for _, r := range s.records {
		list = append(list, r)
	}
	return list
}

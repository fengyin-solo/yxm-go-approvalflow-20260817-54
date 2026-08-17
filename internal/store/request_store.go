package store

import (
	"approvalflow/internal/model"
)

func (s *MemoryStore) CreateRequest(r *model.ApprovalRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.requests {
		if exist.SerialNo == r.SerialNo {
			return ErrConflict
		}
	}
	s.requests[r.ID] = r
	return nil
}

func (s *MemoryStore) GetRequest(id string) (*model.ApprovalRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.requests[id]
	if !ok {
		return nil, ErrNotFound
	}
	return r, nil
}

func (s *MemoryStore) GetRequestBySerialNo(serialNo string) (*model.ApprovalRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.requests {
		if r.SerialNo == serialNo {
			return r, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListRequests() []*model.ApprovalRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.ApprovalRequest, 0, len(s.requests))
	for _, r := range s.requests {
		list = append(list, r)
	}
	return list
}

func (s *MemoryStore) UpdateRequest(r *model.ApprovalRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.requests[r.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.requests {
		if exist.ID != r.ID && exist.SerialNo == r.SerialNo {
			return ErrConflict
		}
	}
	s.requests[r.ID] = r
	return nil
}

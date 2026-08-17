package store

import (
	"approvalflow/internal/model"
)

func (s *MemoryStore) CreateNode(n *model.ApprovalNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.nodes {
		if exist.Seq == n.Seq {
			return ErrConflict
		}
	}
	s.nodes[n.ID] = n
	return nil
}

func (s *MemoryStore) GetNode(id string) (*model.ApprovalNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.nodes[id]
	if !ok {
		return nil, ErrNotFound
	}
	return n, nil
}

func (s *MemoryStore) ListNodes() []*model.ApprovalNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.ApprovalNode, 0, len(s.nodes))
	for _, n := range s.nodes {
		list = append(list, n)
	}
	return list
}

func (s *MemoryStore) DeleteNode(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.nodes[id]; !ok {
		return ErrNotFound
	}
	delete(s.nodes, id)
	return nil
}

// DeleteNodesByTemplate 删除模板下全部节点，返回删除数量。
func (s *MemoryStore) DeleteNodesByTemplate(templateID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	deleted := 0
	for id, n := range s.nodes {
		if n.TemplateID == templateID {
			delete(s.nodes, id)
			deleted++
		}
	}
	return deleted
}

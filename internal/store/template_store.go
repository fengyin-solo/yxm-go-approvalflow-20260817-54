package store

import (
	"approvalflow/internal/model"
)

func (s *MemoryStore) CreateTemplate(t *model.ApprovalTemplate) error {
	for _, exist := range s.templates {
		if exist.Code == t.Code {
			return ErrConflict
		}
	}
	s.templates[t.ID] = t
	return nil
}

func (s *MemoryStore) GetTemplate(id string) (*model.ApprovalTemplate, error) {
	t, ok := s.templates[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (s *MemoryStore) GetTemplateByCode(code string) (*model.ApprovalTemplate, error) {
	for _, t := range s.templates {
		if t.Code == code {
			return t, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListTemplates() []*model.ApprovalTemplate {
	list := make([]*model.ApprovalTemplate, 0, len(s.templates))
	for _, t := range s.templates {
		list = append(list, t)
	}
	return list
}

func (s *MemoryStore) UpdateTemplate(t *model.ApprovalTemplate) error {
	if _, ok := s.templates[t.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.templates {
		if exist.ID != t.ID && exist.Code == t.Code {
			return ErrConflict
		}
	}
	s.templates[t.ID] = t
	return nil
}

func (s *MemoryStore) DeleteTemplate(id string) error {
	if _, ok := s.templates[id]; !ok {
		return ErrNotFound
	}
	delete(s.templates, id)
	return nil
}

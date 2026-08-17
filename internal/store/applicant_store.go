package store

import (
	"approvalflow/internal/model"
)

func (s *MemoryStore) CreateApplicant(a *model.Applicant) error {
	for _, exist := range s.applicants {
		if exist.EmployeeNo == a.EmployeeNo {
			return ErrConflict
		}
	}
	s.applicants[a.ID] = a
	return nil
}

func (s *MemoryStore) GetApplicant(id string) (*model.Applicant, error) {
	a, ok := s.applicants[id]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (s *MemoryStore) GetApplicantByEmployeeNo(no string) (*model.Applicant, error) {
	for _, a := range s.applicants {
		if a.EmployeeNo == no {
			return a, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListApplicants() []*model.Applicant {
	list := make([]*model.Applicant, 0, len(s.applicants))
	for _, a := range s.applicants {
		list = append(list, a)
	}
	return list
}

func (s *MemoryStore) UpdateApplicant(a *model.Applicant) error {
	if _, ok := s.applicants[a.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.applicants {
		if exist.ID != a.ID && exist.EmployeeNo == a.EmployeeNo {
			return ErrConflict
		}
	}
	s.applicants[a.ID] = a
	return nil
}

func (s *MemoryStore) DeleteApplicant(id string) error {
	if _, ok := s.applicants[id]; !ok {
		return ErrNotFound
	}
	delete(s.applicants, id)
	return nil
}

package service

import (
	"sort"
	"time"

	"approvalflow/internal/model"
	"approvalflow/pkg/idgen"
)

// CreateApplicant 创建申请人。
func (s *Service) CreateApplicant(input model.Applicant) (*model.Applicant, error) {
	input.ID = ""
	if err := input.Validate(); err != nil {
		return nil, err
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.CreatedAt = now
	input.UpdatedAt = now
	if err := s.store.CreateApplicant(&input); err != nil {
		return nil, err
	}
	s.log.Infof("创建申请人 %s(%s)", input.Name, input.EmployeeNo)
	return &input, nil
}

// GetApplicant 查询申请人详情。
func (s *Service) GetApplicant(id string) (*model.Applicant, error) {
	return s.store.GetApplicant(id)
}

// ListApplicants 分页查询申请人列表。
func (s *Service) ListApplicants(filter model.ApplicantFilter, page, size int) ([]*model.Applicant, int, error) {
	all := s.store.ListApplicants()
	matched := make([]*model.Applicant, 0, len(all))
	for _, a := range all {
		if filter.Match(a) {
			matched = append(matched, a)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].EmployeeNo < matched[j].EmployeeNo
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Applicant{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateApplicant 更新申请人字段。
func (s *Service) UpdateApplicant(id, name, department, email, status string) (*model.Applicant, error) {
	a, err := s.store.GetApplicant(id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		a.Name = name
	}
	if department != "" {
		a.Department = department
	}
	if email != "" {
		a.Email = email
	}
	if status != "" {
		a.Status = status
	}
	a.UpdatedAt = time.Now()
	if err := a.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateApplicant(a); err != nil {
		return nil, err
	}
	return a, nil
}

// DeleteApplicant 删除申请人；若已发起过审批单或被配置为审批人则拒绝删除。
func (s *Service) DeleteApplicant(id string) error {
	if _, err := s.store.GetApplicant(id); err != nil {
		return err
	}
	for _, r := range s.store.ListRequests() {
		if r.ApplicantID == id {
			return model.NewValidationError("applicant", "该申请人已发起过审批单，不可删除")
		}
	}
	for _, n := range s.store.ListNodes() {
		if n.ApproverID == id {
			return model.NewValidationError("applicant", "该申请人被配置为审批节点审批人，不可删除")
		}
	}
	return s.store.DeleteApplicant(id)
}

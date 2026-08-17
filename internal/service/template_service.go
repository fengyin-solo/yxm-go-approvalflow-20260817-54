package service

import (
	"sort"
	"time"

	"approvalflow/internal/model"
	"approvalflow/internal/store"
	"approvalflow/pkg/idgen"
)

// CreateTemplate 创建审批模板（初始为草稿态）。
func (s *Service) CreateTemplate(input model.ApprovalTemplate) (*model.ApprovalTemplate, error) {
	input.ID = ""
	input.NodeCount = 0
	input.Status = model.TemplateDraft
	if err := input.Validate(); err != nil {
		return nil, err
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.CreatedAt = now
	input.UpdatedAt = now
	if err := s.store.CreateTemplate(&input); err != nil {
		return nil, err
	}
	s.log.Infof("创建审批模板 %s(%s)", input.Name, input.Code)
	return &input, nil
}

// GetTemplate 查询模板详情。
func (s *Service) GetTemplate(id string) (*model.ApprovalTemplate, error) {
	return s.store.GetTemplate(id)
}

// ListTemplates 分页查询模板列表。
func (s *Service) ListTemplates(filter model.TemplateFilter, page, size int) ([]*model.ApprovalTemplate, int, error) {
	all := s.store.ListTemplates()
	matched := make([]*model.ApprovalTemplate, 0, len(all))
	for _, t := range all {
		if filter.Match(t) {
			matched = append(matched, t)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.ApprovalTemplate{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateTemplate 更新模板可编辑字段（已归档模板不可编辑）。
func (s *Service) UpdateTemplate(id, name, category, description string) (*model.ApprovalTemplate, error) {
	t, err := s.store.GetTemplate(id)
	if err != nil {
		return nil, err
	}
	if t.Status == model.TemplateArchived {
		return nil, model.NewValidationError("status", "已归档模板不可编辑")
	}
	if name != "" {
		t.Name = name
	}
	if category != "" {
		t.Category = category
	}
	if description != "" {
		t.Description = description
	}
	t.UpdatedAt = time.Now()
	if err := t.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateTemplate(t); err != nil {
		return nil, err
	}
	return t, nil
}

// TransitionTemplate 模板状态流转（draft->active->archived）。
func (s *Service) TransitionTemplate(id, to string) (*model.ApprovalTemplate, error) {
	t, err := s.store.GetTemplate(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTemplateTransition(t.Status, to) {
		return nil, model.NewValidationError("status", "模板状态不允许从 "+t.Status+" 流转到 "+to)
	}
	// 激活前必须至少配置一个审批节点
	if to == model.TemplateActive {
		nodes := s.nodesOfTemplate(t.ID)
		if len(nodes) == 0 {
			return nil, model.NewValidationError("nodes", "模板未配置审批节点，不能激活")
		}
	}
	t.Status = to
	t.UpdatedAt = time.Now()
	if err := s.store.UpdateTemplate(t); err != nil {
		return nil, err
	}
	s.log.Infof("模板 %s 状态流转为 %s", t.Code, to)
	return t, nil
}

// DeleteTemplate 删除模板；若已有审批单引用则拒绝删除，同时级联删除节点。
func (s *Service) DeleteTemplate(id string) error {
	if _, err := s.store.GetTemplate(id); err != nil {
		return err
	}
	for _, r := range s.store.ListRequests() {
		if r.TemplateID == id {
			return store.ErrConflict
		}
	}
	s.store.DeleteNodesByTemplate(id)
	return s.store.DeleteTemplate(id)
}

// nodesOfTemplate 取模板下按 Seq 升序的节点列表。
func (s *Service) nodesOfTemplate(templateID string) []*model.ApprovalNode {
	all := s.store.ListNodes()
	nodes := make([]*model.ApprovalNode, 0, len(all))
	for _, n := range all {
		if n.TemplateID == templateID {
			nodes = append(nodes, n)
		}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Seq < nodes[j].Seq })
	return nodes
}

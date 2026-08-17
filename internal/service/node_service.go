package service

import (
	"approvalflow/internal/model"
	"approvalflow/pkg/idgen"
	"time"
)

// AddNode 为模板追加审批节点。
// 校验链：模板存在 -> 模板为草稿态 -> 审批人存在 -> Seq 不冲突。
func (s *Service) AddNode(templateID string, seq int, name, approverID string) (*model.ApprovalNode, error) {
	t, err := s.store.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}
	if t.Status != model.TemplateDraft {
		return nil, model.NewValidationError("template", "仅草稿状态的模板可调整节点")
	}
	if _, err := s.store.GetApplicant(approverID); err != nil {
		return nil, model.NewValidationError("approver_id", "审批人不存在")
	}
	node := &model.ApprovalNode{
		ID:         idgen.Hex(),
		TemplateID: templateID,
		Seq:        seq,
		Name:       name,
		ApproverID: approverID,
		CreatedAt:  time.Now(),
	}
	if err := node.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.CreateNode(node); err != nil {
		return nil, err
	}
	t.NodeCount++
	t.UpdatedAt = time.Now()
	if err := s.store.UpdateTemplate(t); err != nil {
		return nil, err
	}
	s.log.Infof("模板 %s 新增节点 #%d %s", t.Code, seq, name)
	return node, nil
}

// ListNodes 查询模板下的节点（按 Seq 升序）。
func (s *Service) ListNodes(templateID string) ([]*model.ApprovalNode, error) {
	if _, err := s.store.GetTemplate(templateID); err != nil {
		return nil, err
	}
	return s.nodesOfTemplate(templateID), nil
}

// RemoveNode 删除草稿模板的某个节点，并同步模板节点计数。
func (s *Service) RemoveNode(nodeID string) error {
	node, err := s.store.GetNode(nodeID)
	if err != nil {
		return err
	}
	t, err := s.store.GetTemplate(node.TemplateID)
	if err != nil {
		return err
	}
	if t.Status != model.TemplateDraft {
		return model.NewValidationError("template", "仅草稿状态的模板可调整节点")
	}
	if err := s.store.DeleteNode(nodeID); err != nil {
		return err
	}
	if t.NodeCount > 0 {
		t.NodeCount--
	}
	t.UpdatedAt = time.Now()
	return s.store.UpdateTemplate(t)
}

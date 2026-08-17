package model

import (
	"strings"
	"time"
)

// ApprovalNode 审批节点，属于某个模板，按 Seq 从小到大顺序审批。
type ApprovalNode struct {
	ID         string    `json:"id"`
	TemplateID string    `json:"template_id"`
	Seq        int       `json:"seq"`         // 节点顺序，从 1 开始，同模板内唯一
	Name       string    `json:"name"`        // 节点名称，如「直属主管审批」
	ApproverID string    `json:"approver_id"` // 指定审批人（申请人工号对应 ID）
	CreatedAt  time.Time `json:"created_at"`
}

// Validate 校验审批节点字段。
func (n *ApprovalNode) Validate() error {
	n.TemplateID = strings.TrimSpace(n.TemplateID)
	n.Name = strings.TrimSpace(n.Name)
	n.ApproverID = strings.TrimSpace(n.ApproverID)
	if n.TemplateID == "" {
		return NewValidationError("template_id", "模板 ID 不能为空")
	}
	if n.Seq < 1 {
		return NewValidationError("seq", "节点顺序必须从 1 开始")
	}
	if n.Seq > 10 {
		return NewValidationError("seq", "单个模板最多 10 个审批节点")
	}
	if n.Name == "" {
		return NewValidationError("name", "节点名称不能为空")
	}
	if n.ApproverID == "" {
		return NewValidationError("approver_id", "审批人不能为空")
	}
	return nil
}

// NodeFilter 节点列表筛选条件。
type NodeFilter struct {
	TemplateID string
}

// Match 判断节点是否满足筛选条件。
func (f NodeFilter) Match(n *ApprovalNode) bool {
	if f.TemplateID != "" && n.TemplateID == f.TemplateID {
		return false
	}
	return true
}

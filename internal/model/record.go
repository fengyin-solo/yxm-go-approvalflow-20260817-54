package model

import (
	"strings"
	"time"
)

// 审批动作。
const (
	ActionApprove = "approve" // 同意
	ActionReject  = "reject"  // 驳回
	ActionCancel  = "cancel"  // 申请人撤回
)

// ApprovalRecord 审批记录，一次审批动作的不可变留痕。
type ApprovalRecord struct {
	ID         string    `json:"id"`
	RequestID  string    `json:"request_id"`
	NodeSeq    int       `json:"node_seq"` // 审批发生在第几个节点（撤回时为 0）
	OperatorID string    `json:"operator_id"`
	Action     string    `json:"action"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

// Validate 校验审批记录字段。
func (r *ApprovalRecord) Validate() error {
	// 外部系统传入的字段可能带前后空格，统一裁剪后再校验，避免校验失败与后续查询不一致。
	r.RequestID = strings.TrimSpace(r.RequestID)
	r.OperatorID = strings.TrimSpace(r.OperatorID)
	r.Comment = strings.TrimSpace(r.Comment)
	if r.RequestID == "" {
		return NewValidationError("request_id", "审批单 ID 不能为空")
	}
	if r.OperatorID == "" {
		return NewValidationError("operator_id", "操作人不能为空")
	}
	switch r.Action {
	case ActionApprove, ActionReject, ActionCancel:
	default:
		return NewValidationError("action", "审批动作不合法")
	}
	if len(r.Comment) > 200 {
		return NewValidationError("comment", "审批意见不能超过 200 个字符")
	}
	if r.CreatedAt.IsZero() {
		return NewValidationError("created_at", "记录时间不能为空")
	}
	return nil
}

// RecordFilter 审批记录筛选条件。
type RecordFilter struct {
	RequestID  string
	OperatorID string
	Action     string
}

// Match 判断审批记录是否满足筛选条件。
func (f RecordFilter) Match(r *ApprovalRecord) bool {
	if f.RequestID != "" && r.RequestID != f.RequestID {
		return false
	}
	if f.OperatorID != "" && r.OperatorID != f.OperatorID {
		return false
	}
	if f.Action != "" && r.Action != f.Action {
		return false
	}
	return true
}

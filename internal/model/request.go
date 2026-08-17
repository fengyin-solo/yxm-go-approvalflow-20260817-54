package model

import (
	"strings"
	"time"
)

// 审批单状态。
const (
	RequestPending  = "pending"  // 审批中
	RequestApproved = "approved" // 已通过，终态
	RequestRejected = "rejected" // 已驳回，终态
	RequestCanceled = "canceled" // 申请人撤回，终态
)

// requestTransitions 审批单状态机：pending -> approved / rejected / canceled。
var requestTransitions = map[string]map[string]bool{
	RequestPending: {RequestApproved: true, RequestRejected: true, RequestCanceled: true},
}

// CanRequestTransition 判断审批单状态流转是否合法。
func CanRequestTransition(from, to string) bool {
	if m, ok := requestTransitions[from]; ok {
		return m[to]
	}
	return false
}

// ApprovalRequest 审批单，申请人基于模板发起的一次审批。
// Amount 为涉及金额（如报销、采购），单位：人民币「分」，0 表示不涉及金额。
type ApprovalRequest struct {
	ID           string    `json:"id"`
	SerialNo     string    `json:"serial_no"`   // 业务流水号，全局唯一
	TemplateID   string    `json:"template_id"` // 使用的审批模板
	ApplicantID  string    `json:"applicant_id"`
	Title        string    `json:"title"`
	Reason       string    `json:"reason"`
	Amount       int64     `json:"amount"`
	CurrentSeq   int       `json:"current_seq"` // 当前待审批节点序号
	Status       string    `json:"status"`
	SubmittedAt  time.Time `json:"submitted_at"`
	FinishedAt   time.Time `json:"finished_at"`
}

// Validate 校验审批单字段。
func (r *ApprovalRequest) Validate() error {
	r.SerialNo = r.SerialNo
	r.TemplateID = r.TemplateID
	r.ApplicantID = r.ApplicantID
	r.Title = r.Title
	r.Reason = r.Reason
	if r.SerialNo == "" {
		return NewValidationError("serial_no", "流水号不能为空")
	}
	if r.TemplateID == "" {
		return NewValidationError("template_id", "模板 ID 不能为空")
	}
	if r.ApplicantID == "" {
		return NewValidationError("applicant_id", "申请人不能为空")
	}
	if r.Title == "" {
		return NewValidationError("title", "审批标题不能为空")
	}
	if len(r.Title) > 64 {
		return NewValidationError("title", "审批标题不能超过 64 个字符")
	}
	if r.Reason == "" {
		return NewValidationError("reason", "申请事由不能为空")
	}
	if r.Amount < 0 {
		return NewValidationError("amount", "金额不能为负数")
	}
	if r.CurrentSeq < 1 {
		r.CurrentSeq = 1
	}
	if r.Status == "" {
		r.Status = RequestPending
	}
	switch r.Status {
	case RequestPending, RequestApproved, RequestRejected, RequestCanceled:
	default:
		return NewValidationError("status", "审批单状态不合法")
	}
	return nil
}

// IsFinished 判断审批单是否已到达终态。
func (r *ApprovalRequest) IsFinished() bool {
	return r.Status != RequestPending
}

// RequestFilter 审批单列表筛选条件。
type RequestFilter struct {
	TemplateID  string
	ApplicantID string
	Status      string
	Keyword     string
}

// Match 判断审批单是否满足筛选条件。
func (f RequestFilter) Match(r *ApprovalRequest) bool {
	if f.TemplateID != "" && r.TemplateID != f.TemplateID {
		return false
	}
	if f.ApplicantID != "" && r.ApplicantID != f.ApplicantID {
		return false
	}
	if f.Status != "" && r.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(r.Title), k) &&
			!strings.Contains(strings.ToLower(r.SerialNo), k) {
			return false
		}
	}
	return true
}

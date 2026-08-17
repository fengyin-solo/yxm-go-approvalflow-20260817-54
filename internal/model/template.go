package model

import (
	"strings"
	"time"
)

// 审批模板状态。
const (
	TemplateDraft    = "draft"    // 草稿，不可用于发起审批
	TemplateActive   = "active"   // 生效中
	TemplateArchived = "archived" // 已归档，终态
)

// templateTransitions 模板状态机：draft -> active -> archived，active 也可直接归档。
var templateTransitions = map[string]map[string]bool{
	TemplateDraft:  {TemplateActive: true},
	TemplateActive: {TemplateArchived: true},
}

// CanTemplateTransition 判断模板状态流转是否合法。
func CanTemplateTransition(from, to string) bool {
	if m, ok := templateTransitions[from]; ok {
		return m[to]
	}
	return false
}

// 审批模板适用的业务类别。
const (
	CategoryLeave    = "leave"    // 请假
	CategoryExpense  = "expense"  // 报销
	CategoryPurchase = "purchase" // 采购
	CategoryGeneral  = "general"  // 通用
)

// ApprovalTemplate 审批模板，定义一类审批的节点编排。
type ApprovalTemplate struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"` // 模板编码，全局唯一
	Name        string    `json:"name"`
	Category    string    `json:"category"`  // leave / expense / purchase / general
	Description string    `json:"description"`
	NodeCount   int       `json:"node_count"` // 关联的审批节点数量（冗余，便于展示）
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate 校验并规范化模板字段。
func (t *ApprovalTemplate) Validate() error {
	// 外部系统传入的字段可能带前后空格，统一裁剪后再校验，避免校验失败与后续查询不一致。
	t.Code = strings.TrimSpace(t.Code)
	t.Name = strings.TrimSpace(t.Name)
	t.Category = strings.TrimSpace(t.Category)
	t.Description = strings.TrimSpace(t.Description)
	if t.Code == "" {
		return NewValidationError("code", "模板编码不能为空")
	}
	if len(t.Code) > 32 {
		return NewValidationError("code", "模板编码不能超过 32 个字符")
	}
	if t.Name == "" {
		return NewValidationError("name", "模板名称不能为空")
	}
	switch t.Category {
	case CategoryLeave, CategoryExpense, CategoryPurchase, CategoryGeneral:
	case "":
		t.Category = CategoryGeneral
	default:
		return NewValidationError("category", "业务类别不合法")
	}
	if t.Status == "" {
		t.Status = TemplateDraft
	}
	switch t.Status {
	case TemplateDraft, TemplateActive, TemplateArchived:
	default:
		return NewValidationError("status", "模板状态不合法")
	}
	return nil
}

// TemplateFilter 模板列表筛选条件。
type TemplateFilter struct {
	Status   string
	Category string
	Keyword  string
}

// Match 判断模板是否满足筛选条件。
func (f TemplateFilter) Match(t *ApprovalTemplate) bool {
	if f.Status != "" && t.Status != f.Status {
		return false
	}
	if f.Category != "" && t.Category != f.Category {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(t.Name), k) &&
			!strings.Contains(strings.ToLower(t.Code), k) {
			return false
		}
	}
	return true
}

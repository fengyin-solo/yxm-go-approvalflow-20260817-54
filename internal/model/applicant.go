package model

import (
	"strings"
	"time"
)

// 申请人状态。
const (
	ApplicantActive   = "active"   // 在职可用
	ApplicantDisabled = "disabled" // 停用（离职等）
)

// Applicant 申请人（员工），发起审批单的主体。
type Applicant struct {
	ID         string    `json:"id"`
	EmployeeNo string    `json:"employee_no"` // 工号，全局唯一
	Name       string    `json:"name"`
	Department string    `json:"department"`
	Email      string    `json:"email"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate 校验并规范化申请人字段。
func (a *Applicant) Validate() error {
	a.EmployeeNo = strings.TrimSpace(a.EmployeeNo)
	a.Name = strings.TrimSpace(a.Name)
	a.Department = strings.TrimSpace(a.Department)
	a.Email = strings.TrimSpace(a.Email)
	if a.EmployeeNo == "" {
		return NewValidationError("employee_no", "工号不能为空")
	}
	if len(a.EmployeeNo) > 20 {
		return NewValidationError("employee_no", "工号不能超过 20 个字符")
	}
	if a.Name == "" {
		return NewValidationError("name", "姓名不能为空")
	}
	if a.Department == "" {
		return NewValidationError("department", "部门不能为空")
	}
	if a.Email != "" && !strings.Contains(a.Email, "@") {
		return NewValidationError("email", "邮箱格式不合法")
	}
	if a.Status == "" {
		a.Status = ApplicantActive
	}
	if a.Status != ApplicantActive && a.Status != ApplicantDisabled {
		return NewValidationError("status", "申请人状态不合法")
	}
	return nil
}

// ApplicantFilter 申请人列表筛选条件。
type ApplicantFilter struct {
	Department string
	Status     string
	Keyword    string
}

// Match 判断申请人是否满足筛选条件。
func (f ApplicantFilter) Match(a *Applicant) bool {
	if f.Department != "" && a.Department != f.Department {
		return false
	}
	if f.Status != "" && a.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(a.Name), k) &&
			!strings.Contains(a.EmployeeNo, k) {
			return false
		}
	}
	return true
}

package service

import (
	"sort"

	"approvalflow/internal/model"
)

// OverviewStats 全局概览统计。
type OverviewStats struct {
	TemplateCount  int   `json:"template_count"`
	ApplicantCount int   `json:"applicant_count"`
	RequestCount   int   `json:"request_count"`
	PendingCount   int   `json:"pending_count"`
	ApprovedCount  int   `json:"approved_count"`
	RejectedCount  int   `json:"rejected_count"`
	CanceledCount  int   `json:"canceled_count"`
	TotalAmount    int64 `json:"total_amount"`    // 全部审批单涉及金额（分）
	ApprovedAmount int64 `json:"approved_amount"` // 已通过审批单金额（分）
	ApprovalRate   int   `json:"approval_rate"`   // 通过率（百分比，四舍五入，基数为已结束的单子）
}

// Stats 全局概览统计。
func (s *Service) Stats() *OverviewStats {
	requests := s.store.ListRequests()
	stats := &OverviewStats{
		TemplateCount:  len(s.store.ListTemplates()),
		ApplicantCount: len(s.store.ListApplicants()),
		RequestCount:   len(requests),
	}
	finished := 0
	for _, r := range requests {
		stats.TotalAmount += r.Amount
		switch r.Status {
		case model.RequestPending:
			stats.PendingCount++
		case model.RequestApproved:
			stats.ApprovedCount++
			stats.ApprovedAmount += r.Amount
			finished++
		case model.RequestRejected:
			stats.RejectedCount++
			finished++
		case model.RequestCanceled:
			stats.CanceledCount++
			finished++
		}
	}
	if finished > 0 {
		stats.ApprovalRate = (stats.ApprovedCount*100 + finished/2) / finished
	}
	return stats
}

// TemplateStat 单模板统计。
type TemplateStat struct {
	TemplateID    string `json:"template_id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	RequestCount  int    `json:"request_count"`
	ApprovedCount int    `json:"approved_count"`
	RejectedCount int    `json:"rejected_count"`
}

// StatsByTemplate 按模板分组统计审批单情况，按单量降序。
func (s *Service) StatsByTemplate() []*TemplateStat {
	byTemplate := make(map[string]*TemplateStat)
	for _, t := range s.store.ListTemplates() {
		byTemplate[t.ID] = &TemplateStat{TemplateID: t.ID, Code: t.Code, Name: t.Name}
	}
	for _, r := range s.store.ListRequests() {
		st, ok := byTemplate[r.TemplateID]
		if !ok {
			continue
		}
		st.RequestCount++
		switch r.Status {
		case model.RequestApproved:
			st.ApprovedCount++
		case model.RequestRejected:
			st.RejectedCount++
		}
	}
	list := make([]*TemplateStat, 0, len(byTemplate))
	for _, st := range byTemplate {
		list = append(list, st)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].RequestCount != list[j].RequestCount {
			return list[i].RequestCount > list[j].RequestCount
		}
		return list[i].Code < list[j].Code
	})
	return list
}

// DepartmentStat 部门统计。
type DepartmentStat struct {
	Department    string `json:"department"`
	RequestCount  int    `json:"request_count"`
	ApprovedCount int    `json:"approved_count"`
	AmountSum     int64  `json:"amount_sum"`
}

// StatsByDepartment 按申请人部门分组统计，按单量降序。
func (s *Service) StatsByDepartment() []*DepartmentStat {
	deptOf := make(map[string]string)
	for _, a := range s.store.ListApplicants() {
		deptOf[a.ID] = a.Department
	}
	byDept := make(map[string]*DepartmentStat)
	for _, r := range s.store.ListRequests() {
		dept := deptOf[r.ApplicantID]
		if dept == "" {
			dept = "未知部门"
		}
		st, ok := byDept[dept]
		if !ok {
			st = &DepartmentStat{Department: dept}
			byDept[dept] = st
		}
		st.RequestCount++
		st.AmountSum += r.Amount
		if r.Status == model.RequestApproved {
			st.ApprovedCount++
		}
	}
	list := make([]*DepartmentStat, 0, len(byDept))
	for _, st := range byDept {
		list = append(list, st)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].RequestCount != list[j].RequestCount {
			return list[i].RequestCount > list[j].RequestCount
		}
		return list[i].Department < list[j].Department
	})
	return list
}

// TopApprover 审批人工作量排行条目。
type TopApprover struct {
	OperatorID   string `json:"operator_id"`
	ApproveCount int    `json:"approve_count"`
	RejectCount  int    `json:"reject_count"`
}

// TopApprovers 按审批动作次数取 TOP N 审批人。
func (s *Service) TopApprovers(n int) []*TopApprover {
	if n <= 0 {
		n = 10
	}
	byOperator := make(map[string]*TopApprover)
	for _, rec := range s.store.ListRecords() {
		st, ok := byOperator[rec.OperatorID]
		if !ok {
			st = &TopApprover{OperatorID: rec.OperatorID}
			byOperator[rec.OperatorID] = st
		}
		switch rec.Action {
		case model.ActionApprove:
			st.ApproveCount++
		case model.ActionReject:
			st.RejectCount++
		}
	}
	list := make([]*TopApprover, 0, len(byOperator))
	for _, st := range byOperator {
		if st.ApproveCount+st.RejectCount == 0 {
			continue
		}
		list = append(list, st)
	}
	sort.Slice(list, func(i, j int) bool {
		ci := list[i].ApproveCount + list[i].RejectCount
		cj := list[j].ApproveCount + list[j].RejectCount
		if ci != cj {
			return ci > cj
		}
		return list[i].OperatorID < list[j].OperatorID
	})
	if len(list) > n {
		list = list[:n]
	}
	return list
}

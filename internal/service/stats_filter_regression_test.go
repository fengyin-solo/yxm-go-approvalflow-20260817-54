package service

import (
	"testing"

	"approvalflow/internal/model"
)

func TestFiltersAndStatsUseStableBusinessFields(t *testing.T) {
	svc := newTestService()
	app, _ := svc.CreateApplicant(model.Applicant{EmployeeNo: "EMP-100", Name: "张三", Department: "财务"})
	approver, _ := svc.CreateApplicant(model.Applicant{EmployeeNo: "EMP-200", Name: "李经理", Department: "财务"})
	tpl, _ := svc.CreateTemplate(model.ApprovalTemplate{Code: "EXPENSE-Q3", Name: "三季度报销", Category: model.CategoryExpense})
	_, _ = svc.AddNode(tpl.ID, 1, "审批", approver.ID)
	_, _ = svc.TransitionTemplate(tpl.ID, model.TemplateActive)

	templates, total, err := svc.ListTemplates(model.TemplateFilter{Keyword: "expense"}, 1, 10)
	if err != nil || total != 1 || templates[0].ID != tpl.ID { t.Fatalf("template keyword should match code case-insensitively, total=%d err=%v", total, err) }
	applicants, total, err := svc.ListApplicants(model.ApplicantFilter{Keyword: "emp-100"}, 1, 10)
	if err != nil || total != 1 || applicants[0].ID != app.ID { t.Fatalf("applicant keyword should match employee no case-insensitively, total=%d err=%v", total, err) }

	r1, _ := svc.SubmitRequest(tpl.ID, app.ID, "差旅报销", "北京出差", 3000)
	r2, _ := svc.SubmitRequest(tpl.ID, app.ID, "采购报销", "设备采购", 7000)
	_, _ = svc.Approve(r1.ID, approver.ID, "ok")
	_, _ = svc.Reject(r2.ID, approver.ID, "预算不足")
	requests, total, err := svc.ListRequests(model.RequestFilter{Keyword: r1.SerialNo[len(r1.SerialNo)-4:]}, 1, 10)
	if err != nil || total != 1 || requests[0].ID != r1.ID { t.Fatalf("request keyword should match serial number, total=%d err=%v", total, err) }

	stats := svc.Stats()
	if stats.TotalAmount != 10000 || stats.ApprovedAmount != 3000 || stats.ApprovalRate != 50 {
		t.Fatalf("unexpected overview stats: %#v", stats)
	}
}

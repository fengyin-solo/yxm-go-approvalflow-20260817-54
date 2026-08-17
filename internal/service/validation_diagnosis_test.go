package service

import (
	"testing"

	"approvalflow/internal/model"
)

func TestRequestValidationAndStateGuardsRemainStrict(t *testing.T) {
	svc := newTestService()
	app, _ := svc.CreateApplicant(model.Applicant{EmployeeNo: "E100", Name: "申请人", Department: "研发"})
	disabled, _ := svc.CreateApplicant(model.Applicant{EmployeeNo: "E999", Name: "离职人员", Department: "研发", Status: model.ApplicantDisabled})
	tpl, _ := svc.CreateTemplate(model.ApprovalTemplate{Code: "STRICT", Name: "严格审批"})
	if _, err := svc.TransitionTemplate(tpl.ID, model.TemplateActive); err == nil {
		t.Fatal("template without nodes must not become active")
	}
	approver, _ := svc.CreateApplicant(model.Applicant{EmployeeNo: "E200", Name: "审批人", Department: "研发"})
	_, _ = svc.AddNode(tpl.ID, 1, "审批", approver.ID)
	_, _ = svc.TransitionTemplate(tpl.ID, model.TemplateActive)
	if _, err := svc.SubmitRequest(tpl.ID, app.ID, "", "有事", 0); err == nil { t.Fatal("empty title must be rejected") }
	if _, err := svc.SubmitRequest(tpl.ID, app.ID, "标题", "", 0); err == nil { t.Fatal("empty reason must be rejected") }
	if _, err := svc.SubmitRequest(tpl.ID, disabled.ID, "标题", "事由", 0); err == nil { t.Fatal("disabled applicant must be rejected") }
}

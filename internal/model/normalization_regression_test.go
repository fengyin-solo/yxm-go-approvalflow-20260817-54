package model

import (
	"testing"
	"time"
)

func TestDomainValidationNormalizesExternalInput(t *testing.T) {
	tpl := &ApprovalTemplate{Code: "  FLOW-A  ", Name: "  报销审批  ", Category: " expense ", Description: "  monthly  "}
	if err := tpl.Validate(); err != nil {
		t.Fatalf("template with padded fields should validate: %v", err)
	}
	if tpl.Code != "FLOW-A" || tpl.Name != "报销审批" || tpl.Category != CategoryExpense || tpl.Description != "monthly" {
		t.Fatalf("template fields were not normalized: %#v", tpl)
	}

	app := &Applicant{EmployeeNo: " E100 ", Name: " 张三 ", Department: " 技术部 ", Email: " zhang@example.com "}
	if err := app.Validate(); err != nil {
		t.Fatalf("applicant with padded fields should validate: %v", err)
	}
	if app.EmployeeNo != "E100" || app.Name != "张三" || app.Department != "技术部" || app.Email != "zhang@example.com" {
		t.Fatalf("applicant fields were not normalized: %#v", app)
	}

	req := &ApprovalRequest{SerialNo: " AP001 ", TemplateID: " tpl-1 ", ApplicantID: " app-1 ", Title: "  年假申请  ", Reason: "  家中有事  "}
	if err := req.Validate(); err != nil {
		t.Fatalf("request with padded fields should validate: %v", err)
	}
	if req.SerialNo != "AP001" || req.TemplateID != "tpl-1" || req.ApplicantID != "app-1" || req.Title != "年假申请" || req.Reason != "家中有事" || req.CurrentSeq != 1 {
		t.Fatalf("request fields were not normalized/defaulted: %#v", req)
	}

	rec := &ApprovalRecord{RequestID: " req-1 ", OperatorID: " mgr-1 ", Action: ActionReject, Comment: "  资料不全  ", CreatedAt: time.Now()}
	if err := rec.Validate(); err != nil {
		t.Fatalf("record with padded fields should validate: %v", err)
	}
	if rec.RequestID != "req-1" || rec.OperatorID != "mgr-1" || rec.Comment != "资料不全" {
		t.Fatalf("record fields were not normalized: %#v", rec)
	}
}

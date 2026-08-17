package service

import (
	"testing"

	"approvalflow/internal/model"
)

func TestApprovalNodesKeepTemplateScopedAscendingOrder(t *testing.T) {
	svc := newTestService()
	a1, _ := svc.CreateApplicant(model.Applicant{EmployeeNo: "A1", Name: "申请人", Department: "研发"})
	m1, _ := svc.CreateApplicant(model.Applicant{EmployeeNo: "M1", Name: "一级", Department: "研发"})
	m2, _ := svc.CreateApplicant(model.Applicant{EmployeeNo: "M2", Name: "二级", Department: "研发"})
	m3, _ := svc.CreateApplicant(model.Applicant{EmployeeNo: "M3", Name: "三级", Department: "研发"})

	t1, _ := svc.CreateTemplate(model.ApprovalTemplate{Code: "FLOW1", Name: "三层审批"})
	t2, _ := svc.CreateTemplate(model.ApprovalTemplate{Code: "FLOW2", Name: "另一模板"})
	if _, err := svc.AddNode(t1.ID, 3, "总经理", m3.ID); err != nil { t.Fatal(err) }
	if _, err := svc.AddNode(t1.ID, 1, "主管", m1.ID); err != nil { t.Fatal(err) }
	if _, err := svc.AddNode(t1.ID, 2, "经理", m2.ID); err != nil { t.Fatal(err) }
	if _, err := svc.AddNode(t2.ID, 1, "另一个主管", m1.ID); err != nil { t.Fatalf("same sequence in another template should be allowed: %v", err) }

	if !((model.NodeFilter{TemplateID: t1.ID}).Match(&model.ApprovalNode{TemplateID: t1.ID})) {
		t.Fatal("node filter should match nodes from the requested template")
	}
	nodes, err := svc.ListNodes(t1.ID)
	if err != nil { t.Fatal(err) }
	if got := []int{nodes[0].Seq, nodes[1].Seq, nodes[2].Seq}; got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("nodes should be sorted ascending, got %v", got)
	}
	t1, err = svc.TransitionTemplate(t1.ID, model.TemplateActive)
	if err != nil { t.Fatal(err) }
	req, err := svc.SubmitRequest(t1.ID, a1.ID, "采购电脑", "项目需要", 1200000)
	if err != nil { t.Fatal(err) }
	if req.CurrentSeq != 1 { t.Fatalf("new request should start at seq=1, got %d", req.CurrentSeq) }
	req, err = svc.Approve(req.ID, m1.ID, "ok")
	if err != nil { t.Fatal(err) }
	if req.CurrentSeq != 2 || req.Status != model.RequestPending { t.Fatalf("after first approval got status=%s seq=%d", req.Status, req.CurrentSeq) }
	req, err = svc.Approve(req.ID, m2.ID, "ok")
	if err != nil { t.Fatal(err) }
	if req.CurrentSeq != 3 || req.Status != model.RequestPending { t.Fatalf("after second approval got status=%s seq=%d", req.Status, req.CurrentSeq) }
}

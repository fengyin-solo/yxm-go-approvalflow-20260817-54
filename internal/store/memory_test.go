package store

import (
	"testing"
	"time"

	"approvalflow/internal/model"
)

func newTemplate(id, code string) *model.ApprovalTemplate {
	return &model.ApprovalTemplate{
		ID:        id,
		Code:      code,
		Name:      "测试模板",
		Category:  model.CategoryGeneral,
		Status:    model.TemplateDraft,
		CreatedAt: time.Now(),
	}
}

func TestTemplateCRUD(t *testing.T) {
	s := NewMemoryStore()
	tpl := newTemplate("t1", "LEAVE")

	if err := s.CreateTemplate(tpl); err != nil {
		t.Fatalf("创建模板失败: %v", err)
	}
	got, err := s.GetTemplate("t1")
	if err != nil || got.Code != "LEAVE" {
		t.Fatalf("查询模板失败: %v", err)
	}
	if _, err := s.GetTemplateByCode("LEAVE"); err != nil {
		t.Fatalf("按编码查询失败: %v", err)
	}
	if _, err := s.GetTemplateByCode("NOPE"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
	if list := s.ListTemplates(); len(list) != 1 {
		t.Fatalf("期望 1 个模板，得到 %d", len(list))
	}
	tpl.Name = "新名字"
	if err := s.UpdateTemplate(tpl); err != nil {
		t.Fatalf("更新模板失败: %v", err)
	}
	if err := s.DeleteTemplate("t1"); err != nil {
		t.Fatalf("删除模板失败: %v", err)
	}
	if _, err := s.GetTemplate("t1"); err != ErrNotFound {
		t.Fatalf("删除后期望 ErrNotFound，得到 %v", err)
	}
}

func TestTemplateCodeConflict(t *testing.T) {
	s := NewMemoryStore()
	if err := s.CreateTemplate(newTemplate("t1", "DUP")); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateTemplate(newTemplate("t2", "DUP")); err != ErrConflict {
		t.Fatalf("期望 ErrConflict，得到 %v", err)
	}
	other := newTemplate("t3", "OTHER")
	_ = s.CreateTemplate(other)
	other.Code = "DUP"
	if err := s.UpdateTemplate(other); err != ErrConflict {
		t.Fatalf("更新编码冲突期望 ErrConflict，得到 %v", err)
	}
}

func TestTemplateNotFoundOps(t *testing.T) {
	s := NewMemoryStore()
	if _, err := s.GetTemplate("missing"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
	if err := s.UpdateTemplate(newTemplate("missing", "X")); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
	if err := s.DeleteTemplate("missing"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
}

func newApplicant(id, no string) *model.Applicant {
	return &model.Applicant{
		ID:         id,
		EmployeeNo: no,
		Name:       "员工" + no,
		Department: "技术部",
		Status:     model.ApplicantActive,
		CreatedAt:  time.Now(),
	}
}

func TestApplicantCRUDAndConflict(t *testing.T) {
	s := NewMemoryStore()
	a := newApplicant("a1", "E001")
	if err := s.CreateApplicant(a); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateApplicant(newApplicant("a2", "E001")); err != ErrConflict {
		t.Fatalf("重复工号期望 ErrConflict，得到 %v", err)
	}
	if got, err := s.GetApplicantByEmployeeNo("E001"); err != nil || got.ID != "a1" {
		t.Fatalf("按工号查询失败: %v", err)
	}
	if _, err := s.GetApplicantByEmployeeNo("NOPE"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
	if list := s.ListApplicants(); len(list) != 1 {
		t.Fatalf("期望 1 个申请人，得到 %d", len(list))
	}
	a.Department = "产品部"
	if err := s.UpdateApplicant(a); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteApplicant("a1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetApplicant("a1"); err != ErrNotFound {
		t.Fatalf("删除后期望 ErrNotFound，得到 %v", err)
	}
}

func newNode(id, templateID string, seq int) *model.ApprovalNode {
	return &model.ApprovalNode{
		ID:         id,
		TemplateID: templateID,
		Seq:        seq,
		Name:       "节点",
		ApproverID: "a1",
		CreatedAt:  time.Now(),
	}
}

func TestNodeCRUDAndSeqConflict(t *testing.T) {
	s := NewMemoryStore()
	n := newNode("n1", "t1", 1)
	if err := s.CreateNode(n); err != nil {
		t.Fatal(err)
	}
	// 同模板同 Seq 冲突
	if err := s.CreateNode(newNode("n2", "t1", 1)); err != ErrConflict {
		t.Fatalf("同模板同序号期望 ErrConflict，得到 %v", err)
	}
	// 不同模板同 Seq 不冲突
	if err := s.CreateNode(newNode("n3", "t2", 1)); err != nil {
		t.Fatalf("不同模板同序号不应冲突: %v", err)
	}
	if got, err := s.GetNode("n1"); err != nil || got.Seq != 1 {
		t.Fatalf("查询节点失败: %v", err)
	}
	if list := s.ListNodes(); len(list) != 2 {
		t.Fatalf("期望 2 个节点，得到 %d", len(list))
	}
	if err := s.DeleteNode("n1"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteNode("n1"); err != ErrNotFound {
		t.Fatalf("重复删除期望 ErrNotFound，得到 %v", err)
	}
}

func TestDeleteNodesByTemplate(t *testing.T) {
	s := NewMemoryStore()
	_ = s.CreateNode(newNode("n1", "t1", 1))
	_ = s.CreateNode(newNode("n2", "t1", 2))
	_ = s.CreateNode(newNode("n3", "t2", 1))
	if got := s.DeleteNodesByTemplate("t1"); got != 2 {
		t.Fatalf("期望删除 2 个，得到 %d", got)
	}
	if list := s.ListNodes(); len(list) != 1 {
		t.Fatalf("期望剩 1 个节点，得到 %d", len(list))
	}
	if got := s.DeleteNodesByTemplate("t9"); got != 0 {
		t.Fatalf("无匹配期望 0，得到 %d", got)
	}
}

func newRequest(id, serialNo string) *model.ApprovalRequest {
	return &model.ApprovalRequest{
		ID:          id,
		SerialNo:    serialNo,
		TemplateID:  "t1",
		ApplicantID: "a1",
		Title:       "测试审批",
		Reason:      "事由",
		CurrentSeq:  1,
		Status:      model.RequestPending,
		SubmittedAt: time.Now(),
	}
}

func TestRequestCRUDAndConflict(t *testing.T) {
	s := NewMemoryStore()
	r := newRequest("r1", "AP001")
	if err := s.CreateRequest(r); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRequest(newRequest("r2", "AP001")); err != ErrConflict {
		t.Fatalf("重复流水号期望 ErrConflict，得到 %v", err)
	}
	if got, err := s.GetRequestBySerialNo("AP001"); err != nil || got.ID != "r1" {
		t.Fatalf("按流水号查询失败: %v", err)
	}
	if _, err := s.GetRequestBySerialNo("NOPE"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
	if list := s.ListRequests(); len(list) != 1 {
		t.Fatalf("期望 1 个审批单，得到 %d", len(list))
	}
	r.Status = model.RequestApproved
	if err := s.UpdateRequest(r); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetRequest("r1"); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateRequest(newRequest("missing", "APX")); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
}

func newRecord(id, requestID string) *model.ApprovalRecord {
	return &model.ApprovalRecord{
		ID:         id,
		RequestID:  requestID,
		NodeSeq:    1,
		OperatorID: "a2",
		Action:     model.ActionApprove,
		CreatedAt:  time.Now(),
	}
}

func TestRecordCreateAndList(t *testing.T) {
	s := NewMemoryStore()
	if err := s.CreateRecord(newRecord("rec1", "r1")); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRecord(newRecord("rec2", "r1")); err != nil {
		t.Fatal(err)
	}
	if got, err := s.GetRecord("rec1"); err != nil || got.RequestID != "r1" {
		t.Fatalf("查询记录失败: %v", err)
	}
	if _, err := s.GetRecord("missing"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
	if list := s.ListRecords(); len(list) != 2 {
		t.Fatalf("期望 2 条记录，得到 %d", len(list))
	}
}

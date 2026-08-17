package model

import (
	"testing"
	"time"
)

func TestTemplateValidate(t *testing.T) {
	base := func() *ApprovalTemplate {
		return &ApprovalTemplate{Code: "T1", Name: "模板", Category: CategoryLeave}
	}
	cases := []struct {
		name    string
		mutate  func(*ApprovalTemplate)
		wantErr bool
	}{
		{"合法默认值", func(t *ApprovalTemplate) {}, false},
		{"空编码", func(t *ApprovalTemplate) { t.Code = "  " }, true},
		{"编码过长", func(t *ApprovalTemplate) { t.Code = "012345678901234567890123456789012345" }, true},
		{"空名称", func(t *ApprovalTemplate) { t.Name = "" }, true},
		{"非法类别", func(t *ApprovalTemplate) { t.Category = "xxx" }, true},
		{"空类别回落 general", func(t *ApprovalTemplate) { t.Category = "" }, false},
		{"非法状态", func(t *ApprovalTemplate) { t.Status = "weird" }, true},
	}
	for _, c := range cases {
		tpl := base()
		c.mutate(tpl)
		err := tpl.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
	// 空类别应回落为 general
	tpl := base()
	tpl.Category = ""
	_ = tpl.Validate()
	if tpl.Category != CategoryGeneral {
		t.Errorf("空类别应回落为 general，得到 %s", tpl.Category)
	}
}

func TestTemplateTransition(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{TemplateDraft, TemplateActive, true},
		{TemplateDraft, TemplateArchived, false},
		{TemplateActive, TemplateArchived, true},
		{TemplateActive, TemplateDraft, false},
		{TemplateArchived, TemplateActive, false},
		{"unknown", TemplateActive, false},
	}
	for _, c := range cases {
		if got := CanTemplateTransition(c.from, c.to); got != c.want {
			t.Errorf("CanTemplateTransition(%s,%s)=%v，期望 %v", c.from, c.to, got, c.want)
		}
	}
}

func TestTemplateFilterMatch(t *testing.T) {
	tpl := &ApprovalTemplate{Code: "LEAVE", Name: "请假审批", Category: CategoryLeave, Status: TemplateActive}
	if !(TemplateFilter{}).Match(tpl) {
		t.Error("空筛选应匹配所有")
	}
	if !(TemplateFilter{Status: TemplateActive, Category: CategoryLeave}).Match(tpl) {
		t.Error("状态+类别匹配应通过")
	}
	if (TemplateFilter{Status: TemplateDraft}).Match(tpl) {
		t.Error("状态不匹配应失败")
	}
	if !(TemplateFilter{Keyword: "请假"}).Match(tpl) {
		t.Error("关键词匹配名称应通过")
	}
	if !(TemplateFilter{Keyword: "leave"}).Match(tpl) {
		t.Error("关键词匹配编码（忽略大小写）应通过")
	}
	if (TemplateFilter{Keyword: "报销"}).Match(tpl) {
		t.Error("关键词不匹配应失败")
	}
}

func TestApplicantValidate(t *testing.T) {
	base := func() *Applicant {
		return &Applicant{EmployeeNo: "E1", Name: "张三", Department: "技术部"}
	}
	cases := []struct {
		name    string
		mutate  func(*Applicant)
		wantErr bool
	}{
		{"合法", func(a *Applicant) {}, false},
		{"空工号", func(a *Applicant) { a.EmployeeNo = "" }, true},
		{"工号过长", func(a *Applicant) { a.EmployeeNo = "E123456789012345678901" }, true},
		{"空姓名", func(a *Applicant) { a.Name = "" }, true},
		{"空部门", func(a *Applicant) { a.Department = "" }, true},
		{"非法邮箱", func(a *Applicant) { a.Email = "no-at" }, true},
		{"合法邮箱", func(a *Applicant) { a.Email = "a@b.com" }, false},
		{"非法状态", func(a *Applicant) { a.Status = "weird" }, true},
	}
	for _, c := range cases {
		a := base()
		c.mutate(a)
		err := a.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
}

func TestNodeValidate(t *testing.T) {
	base := func() *ApprovalNode {
		return &ApprovalNode{TemplateID: "t1", Seq: 1, Name: "审批", ApproverID: "a1"}
	}
	cases := []struct {
		name    string
		mutate  func(*ApprovalNode)
		wantErr bool
	}{
		{"合法", func(n *ApprovalNode) {}, false},
		{"空模板", func(n *ApprovalNode) { n.TemplateID = "" }, true},
		{"Seq 为 0", func(n *ApprovalNode) { n.Seq = 0 }, true},
		{"Seq 超上限", func(n *ApprovalNode) { n.Seq = 11 }, true},
		{"空名称", func(n *ApprovalNode) { n.Name = "" }, true},
		{"空审批人", func(n *ApprovalNode) { n.ApproverID = "" }, true},
	}
	for _, c := range cases {
		n := base()
		c.mutate(n)
		err := n.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
}

func TestRequestValidateAndTransition(t *testing.T) {
	base := func() *ApprovalRequest {
		return &ApprovalRequest{
			SerialNo: "AP1", TemplateID: "t1", ApplicantID: "a1",
			Title: "标题", Reason: "事由",
		}
	}
	cases := []struct {
		name    string
		mutate  func(*ApprovalRequest)
		wantErr bool
	}{
		{"合法", func(r *ApprovalRequest) {}, false},
		{"空流水号", func(r *ApprovalRequest) { r.SerialNo = "" }, true},
		{"空模板", func(r *ApprovalRequest) { r.TemplateID = "" }, true},
		{"空申请人", func(r *ApprovalRequest) { r.ApplicantID = "" }, true},
		{"空标题", func(r *ApprovalRequest) { r.Title = "" }, true},
		{"标题过长", func(r *ApprovalRequest) { r.Title = string(make([]byte, 65)) }, true},
		{"空事由", func(r *ApprovalRequest) { r.Reason = "" }, true},
		{"负金额", func(r *ApprovalRequest) { r.Amount = -1 }, true},
		{"非法状态", func(r *ApprovalRequest) { r.Status = "weird" }, true},
	}
	for _, c := range cases {
		r := base()
		c.mutate(r)
		err := r.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
	// 状态机
	transitions := []struct {
		from, to string
		want     bool
	}{
		{RequestPending, RequestApproved, true},
		{RequestPending, RequestRejected, true},
		{RequestPending, RequestCanceled, true},
		{RequestApproved, RequestRejected, false},
		{RequestRejected, RequestApproved, false},
		{RequestCanceled, RequestPending, false},
	}
	for _, c := range transitions {
		if got := CanRequestTransition(c.from, c.to); got != c.want {
			t.Errorf("CanRequestTransition(%s,%s)=%v，期望 %v", c.from, c.to, got, c.want)
		}
	}
	// IsFinished
	r := base()
	r.Status = RequestPending
	if r.IsFinished() {
		t.Error("pending 不应为已结束")
	}
	r.Status = RequestApproved
	if !r.IsFinished() {
		t.Error("approved 应为已结束")
	}
}

func TestRecordValidate(t *testing.T) {
	base := func() *ApprovalRecord {
		return &ApprovalRecord{
			RequestID: "r1", OperatorID: "a1", Action: ActionApprove, CreatedAt: time.Now(),
		}
	}
	cases := []struct {
		name    string
		mutate  func(*ApprovalRecord)
		wantErr bool
	}{
		{"合法", func(r *ApprovalRecord) {}, false},
		{"空审批单", func(r *ApprovalRecord) { r.RequestID = "" }, true},
		{"空操作人", func(r *ApprovalRecord) { r.OperatorID = "" }, true},
		{"非法动作", func(r *ApprovalRecord) { r.Action = "weird" }, true},
		{"意见过长", func(r *ApprovalRecord) { r.Comment = string(make([]byte, 201)) }, true},
		{"零值时间", func(r *ApprovalRecord) { r.CreatedAt = time.Time{} }, true},
	}
	for _, c := range cases {
		r := base()
		c.mutate(r)
		err := r.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
}

func TestRequestFilterMatch(t *testing.T) {
	r := &ApprovalRequest{
		SerialNo: "AP100", TemplateID: "t1", ApplicantID: "a1",
		Title: "报销审批", Status: RequestPending,
	}
	if !(RequestFilter{}).Match(r) {
		t.Error("空筛选应匹配所有")
	}
	if !(RequestFilter{TemplateID: "t1", ApplicantID: "a1", Status: RequestPending}).Match(r) {
		t.Error("全条件匹配应通过")
	}
	if (RequestFilter{Status: RequestApproved}).Match(r) {
		t.Error("状态不匹配应失败")
	}
	if !(RequestFilter{Keyword: "报销"}).Match(r) {
		t.Error("关键词匹配标题应通过")
	}
	if !(RequestFilter{Keyword: "ap100"}).Match(r) {
		t.Error("关键词匹配流水号（忽略大小写）应通过")
	}
}

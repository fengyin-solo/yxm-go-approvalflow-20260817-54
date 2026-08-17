package service

import (
	"testing"

	"approvalflow/internal/config"
	"approvalflow/internal/model"
	"approvalflow/internal/store"
	"approvalflow/pkg/logger"
)

func newTestService() *Service {
	cfg := &config.Config{MaxPageSize: 100}
	log := logger.NewLevel(logger.LevelError)
	return New(store.NewMemoryStore(), log, cfg)
}

// setupTwoLevelTemplate 创建申请人、两级审批模板并激活，返回 (模板, 申请人, 审批人1, 审批人2)。
func setupTwoLevelTemplate(t *testing.T, svc *Service) (*model.ApprovalTemplate, *model.Applicant, *model.Applicant, *model.Applicant) {
	t.Helper()
	applicant, err := svc.CreateApplicant(model.Applicant{
		EmployeeNo: "E100", Name: "张三", Department: "技术部",
	})
	if err != nil {
		t.Fatal(err)
	}
	approver1, err := svc.CreateApplicant(model.Applicant{
		EmployeeNo: "E200", Name: "李主管", Department: "技术部",
	})
	if err != nil {
		t.Fatal(err)
	}
	approver2, err := svc.CreateApplicant(model.Applicant{
		EmployeeNo: "E300", Name: "王总监", Department: "技术部",
	})
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := svc.CreateTemplate(model.ApprovalTemplate{
		Code: "LEAVE2", Name: "请假两级审批", Category: model.CategoryLeave,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddNode(tpl.ID, 1, "直属主管审批", approver1.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddNode(tpl.ID, 2, "总监审批", approver2.ID); err != nil {
		t.Fatal(err)
	}
	tpl, err = svc.TransitionTemplate(tpl.ID, model.TemplateActive)
	if err != nil {
		t.Fatal(err)
	}
	return tpl, applicant, approver1, approver2
}

func TestCreateTemplateValidation(t *testing.T) {
	svc := newTestService()
	cases := []struct {
		name  string
		input model.ApprovalTemplate
	}{
		{"空编码", model.ApprovalTemplate{Name: "x"}},
		{"空名称", model.ApprovalTemplate{Code: "C1"}},
		{"非法类别", model.ApprovalTemplate{Code: "C1", Name: "x", Category: "weird"}},
	}
	for _, c := range cases {
		if _, err := svc.CreateTemplate(c.input); err == nil {
			t.Errorf("%s: 期望校验失败但成功了", c.name)
		} else if !model.IsValidationError(err) {
			t.Errorf("%s: 期望 ValidationError，得到 %v", c.name, err)
		}
	}
}

func TestTemplateLifecycleAndActivation(t *testing.T) {
	svc := newTestService()
	tpl, err := svc.CreateTemplate(model.ApprovalTemplate{Code: "LIFE", Name: "生命周期"})
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Status != model.TemplateDraft {
		t.Fatalf("新模板应为 draft，得到 %s", tpl.Status)
	}
	// 无节点不能激活
	if _, err := svc.TransitionTemplate(tpl.ID, model.TemplateActive); err == nil {
		t.Fatal("无节点模板激活应被拒绝")
	}
	// 非法流转 draft -> archived
	if _, err := svc.TransitionTemplate(tpl.ID, model.TemplateArchived); err == nil {
		t.Fatal("draft->archived 应被拒绝")
	}
	// 加节点后可激活
	a, _ := svc.CreateApplicant(model.Applicant{EmployeeNo: "E1", Name: "甲", Department: "d"})
	if _, err := svc.AddNode(tpl.ID, 1, "审批", a.ID); err != nil {
		t.Fatal(err)
	}
	tpl, err = svc.TransitionTemplate(tpl.ID, model.TemplateActive)
	if err != nil || tpl.Status != model.TemplateActive {
		t.Fatalf("激活失败: %v", err)
	}
	// active -> archived
	tpl, err = svc.TransitionTemplate(tpl.ID, model.TemplateArchived)
	if err != nil || tpl.Status != model.TemplateArchived {
		t.Fatalf("归档失败: %v", err)
	}
	// 归档后不可编辑
	if _, err := svc.UpdateTemplate(tpl.ID, "新名字", "", ""); err == nil {
		t.Fatal("归档模板编辑应被拒绝")
	}
}

func TestNodeManagement(t *testing.T) {
	svc := newTestService()
	tpl, _ := svc.CreateTemplate(model.ApprovalTemplate{Code: "NODE", Name: "节点管理"})
	a, _ := svc.CreateApplicant(model.Applicant{EmployeeNo: "E1", Name: "甲", Department: "d"})

	// 审批人不存在
	if _, err := svc.AddNode(tpl.ID, 1, "审批", "missing"); err == nil {
		t.Fatal("审批人不存在应被拒绝")
	}
	// Seq 非法
	if _, err := svc.AddNode(tpl.ID, 0, "审批", a.ID); err == nil {
		t.Fatal("Seq=0 应被拒绝")
	}
	n1, err := svc.AddNode(tpl.ID, 1, "一级审批", a.ID)
	if err != nil {
		t.Fatal(err)
	}
	// 同 Seq 冲突
	if _, err := svc.AddNode(tpl.ID, 1, "重复", a.ID); err == nil {
		t.Fatal("同序号节点应冲突")
	}
	nodes, err := svc.ListNodes(tpl.ID)
	if err != nil || len(nodes) != 1 {
		t.Fatalf("期望 1 个节点，得到 %d err=%v", len(nodes), err)
	}
	got, _ := svc.GetTemplate(tpl.ID)
	if got.NodeCount != 1 {
		t.Fatalf("模板节点计数应为 1，得到 %d", got.NodeCount)
	}
	// 删除节点
	if err := svc.RemoveNode(n1.ID); err != nil {
		t.Fatal(err)
	}
	got, _ = svc.GetTemplate(tpl.ID)
	if got.NodeCount != 0 {
		t.Fatalf("删除后节点计数应为 0，得到 %d", got.NodeCount)
	}
}

func TestDeleteTemplateWithRequests(t *testing.T) {
	svc := newTestService()
	tpl, applicant, _, _ := setupTwoLevelTemplate(t, svc)
	if _, err := svc.SubmitRequest(tpl.ID, applicant.ID, "请假", "家中有事", 0); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteTemplate(tpl.ID); err == nil {
		t.Fatal("已有审批单的模板删除应被拒绝")
	}
}

func TestSubmitRequestValidation(t *testing.T) {
	svc := newTestService()
	tpl, applicant, _, _ := setupTwoLevelTemplate(t, svc)

	// 草稿模板不能发起
	draft, _ := svc.CreateTemplate(model.ApprovalTemplate{Code: "DRAFT", Name: "草稿"})
	if _, err := svc.SubmitRequest(draft.ID, applicant.ID, "x", "y", 0); err == nil {
		t.Fatal("草稿模板发起审批应被拒绝")
	}
	// 申请人不存在
	if _, err := svc.SubmitRequest(tpl.ID, "missing", "x", "y", 0); err == nil {
		t.Fatal("申请人不存在应被拒绝")
	}
	// 空标题
	if _, err := svc.SubmitRequest(tpl.ID, applicant.ID, "", "y", 0); err == nil {
		t.Fatal("空标题应被拒绝")
	}
	// 负金额
	if _, err := svc.SubmitRequest(tpl.ID, applicant.ID, "x", "y", -1); err == nil {
		t.Fatal("负金额应被拒绝")
	}
	// 停用申请人不能发起
	disabled, _ := svc.CreateApplicant(model.Applicant{
		EmployeeNo: "E999", Name: "离职", Department: "d", Status: model.ApplicantDisabled,
	})
	if _, err := svc.SubmitRequest(tpl.ID, disabled.ID, "x", "y", 0); err == nil {
		t.Fatal("停用申请人发起审批应被拒绝")
	}
}

func TestFullApprovalFlow(t *testing.T) {
	svc := newTestService()
	tpl, applicant, approver1, approver2 := setupTwoLevelTemplate(t, svc)

	req, err := svc.SubmitRequest(tpl.ID, applicant.ID, "年假 3 天", "家中有事", 0)
	if err != nil {
		t.Fatal(err)
	}
	if req.Status != model.RequestPending || req.CurrentSeq != 1 {
		t.Fatalf("初始状态不对: %s seq=%d", req.Status, req.CurrentSeq)
	}
	// 非当前节点审批人不能审批
	if _, err := svc.Approve(req.ID, approver2.ID, "ok"); err == nil {
		t.Fatal("第二节点审批人越级审批应被拒绝")
	}
	// 第一节点通过，流转到第二节点
	req, err = svc.Approve(req.ID, approver1.ID, "同意")
	if err != nil {
		t.Fatal(err)
	}
	if req.Status != model.RequestPending || req.CurrentSeq != 2 {
		t.Fatalf("第一节点通过后应流转到 seq=2，得到 %s seq=%d", req.Status, req.CurrentSeq)
	}
	// 第二节点通过，审批单完成
	req, err = svc.Approve(req.ID, approver2.ID, "批准")
	if err != nil {
		t.Fatal(err)
	}
	if req.Status != model.RequestApproved {
		t.Fatalf("全部通过后应为 approved，得到 %s", req.Status)
	}
	if req.FinishedAt.IsZero() {
		t.Fatal("完成后应设置 FinishedAt")
	}
	// 已结束的单子不能再审批
	if _, err := svc.Approve(req.ID, approver2.ID, "再来一次"); err == nil {
		t.Fatal("已结束审批单再审批应被拒绝")
	}
	// 时间线应有 2 条记录
	records, err := svc.Timeline(req.ID)
	if err != nil || len(records) != 2 {
		t.Fatalf("期望 2 条时间线记录，得到 %d err=%v", len(records), err)
	}
}

func TestRejectFlow(t *testing.T) {
	svc := newTestService()
	tpl, applicant, approver1, _ := setupTwoLevelTemplate(t, svc)
	req, err := svc.SubmitRequest(tpl.ID, applicant.ID, "报销", "出差", 50000)
	if err != nil {
		t.Fatal(err)
	}
	// 驳回必须填意见
	if _, err := svc.Reject(req.ID, approver1.ID, ""); err == nil {
		t.Fatal("驳回不填意见应被拒绝")
	}
	req, err = svc.Reject(req.ID, approver1.ID, "发票不全")
	if err != nil {
		t.Fatal(err)
	}
	if req.Status != model.RequestRejected {
		t.Fatalf("驳回后应为 rejected，得到 %s", req.Status)
	}
}

func TestCancelFlow(t *testing.T) {
	svc := newTestService()
	tpl, applicant, approver1, _ := setupTwoLevelTemplate(t, svc)
	req, err := svc.SubmitRequest(tpl.ID, applicant.ID, "请假", "私事", 0)
	if err != nil {
		t.Fatal(err)
	}
	// 非申请人不能撤回
	if _, err := svc.Cancel(req.ID, approver1.ID, "撤回"); err == nil {
		t.Fatal("非申请人撤回应被拒绝")
	}
	req, err = svc.Cancel(req.ID, applicant.ID, "不需要了")
	if err != nil {
		t.Fatal(err)
	}
	if req.Status != model.RequestCanceled {
		t.Fatalf("撤回后应为 canceled，得到 %s", req.Status)
	}
	// 已撤回不能再撤回
	if _, err := svc.Cancel(req.ID, applicant.ID, "再撤"); err == nil {
		t.Fatal("已撤回单子再撤回应被拒绝")
	}
}

func TestPendingOf(t *testing.T) {
	svc := newTestService()
	tpl, applicant, approver1, approver2 := setupTwoLevelTemplate(t, svc)
	req1, _ := svc.SubmitRequest(tpl.ID, applicant.ID, "单1", "a", 0)
	req2, _ := svc.SubmitRequest(tpl.ID, applicant.ID, "单2", "b", 0)

	pending, err := svc.PendingOf(approver1.ID)
	if err != nil || len(pending) != 2 {
		t.Fatalf("审批人1 应有 2 个待办，得到 %d err=%v", len(pending), err)
	}
	pending, err = svc.PendingOf(approver2.ID)
	if err != nil || len(pending) != 0 {
		t.Fatalf("审批人2 初始应无待办，得到 %d", len(pending))
	}
	// 推进 req1 到第二节点后，审批人2 有 1 个待办
	if _, err := svc.Approve(req1.ID, approver1.ID, "ok"); err != nil {
		t.Fatal(err)
	}
	pending, _ = svc.PendingOf(approver2.ID)
	if len(pending) != 1 || pending[0].ID != req1.ID {
		t.Fatalf("审批人2 应有 req1 待办，得到 %d", len(pending))
	}
	pending, _ = svc.PendingOf(approver1.ID)
	if len(pending) != 1 || pending[0].ID != req2.ID {
		t.Fatalf("审批人1 应剩 req2 待办，得到 %d", len(pending))
	}
	// 不存在的审批人
	if _, err := svc.PendingOf("missing"); err == nil {
		t.Fatal("不存在的审批人查询应报错")
	}
}

func TestApplicantCRUDAndDeleteGuard(t *testing.T) {
	svc := newTestService()
	a, err := svc.CreateApplicant(model.Applicant{
		EmployeeNo: "E500", Name: "赵六", Department: "市场部", Email: "zhao@test.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	// 重复工号
	if _, err := svc.CreateApplicant(model.Applicant{EmployeeNo: "E500", Name: "x", Department: "d"}); err == nil {
		t.Fatal("重复工号应冲突")
	}
	// 非法邮箱
	if _, err := svc.CreateApplicant(model.Applicant{EmployeeNo: "E501", Name: "x", Department: "d", Email: "bad"}); err == nil {
		t.Fatal("非法邮箱应被拒绝")
	}
	// 更新
	updated, err := svc.UpdateApplicant(a.ID, "", "销售部", "", "")
	if err != nil || updated.Department != "销售部" {
		t.Fatalf("更新失败: %v", err)
	}
	// 列表筛选
	items, total, err := svc.ListApplicants(model.ApplicantFilter{Department: "销售部"}, 1, 10)
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("部门筛选失败: total=%d err=%v", total, err)
	}
	// 被配置为审批人不能删除
	tpl, _ := svc.CreateTemplate(model.ApprovalTemplate{Code: "GUARD", Name: "删除保护"})
	if _, err := svc.AddNode(tpl.ID, 1, "审批", a.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteApplicant(a.ID); err == nil {
		t.Fatal("被配置为审批人的申请人删除应被拒绝")
	}
}

func TestStatsOverviewAndGrouping(t *testing.T) {
	svc := newTestService()
	tpl, applicant, approver1, approver2 := setupTwoLevelTemplate(t, svc)

	// 单1：通过
	r1, _ := svc.SubmitRequest(tpl.ID, applicant.ID, "单1", "a", 10000)
	svc.Approve(r1.ID, approver1.ID, "ok")
	svc.Approve(r1.ID, approver2.ID, "ok")
	// 单2：驳回
	r2, _ := svc.SubmitRequest(tpl.ID, applicant.ID, "单2", "b", 20000)
	svc.Reject(r2.ID, approver1.ID, "不行")
	// 单3：撤回
	r3, _ := svc.SubmitRequest(tpl.ID, applicant.ID, "单3", "c", 30000)
	svc.Cancel(r3.ID, applicant.ID, "撤回")
	// 单4：pending
	svc.SubmitRequest(tpl.ID, applicant.ID, "单4", "d", 40000)

	overview := svc.Stats()
	if overview.RequestCount != 4 || overview.PendingCount != 1 {
		t.Fatalf("概览数量不对: req=%d pending=%d", overview.RequestCount, overview.PendingCount)
	}
	if overview.ApprovedCount != 1 || overview.RejectedCount != 1 || overview.CanceledCount != 1 {
		t.Fatalf("终态计数不对: %+v", overview)
	}
	if overview.TotalAmount != 100000 {
		t.Fatalf("总金额期望 100000，得到 %d", overview.TotalAmount)
	}
	if overview.ApprovedAmount != 10000 {
		t.Fatalf("通过金额期望 10000，得到 %d", overview.ApprovedAmount)
	}
	// 通过率：已结束 3 单，通过 1 单 = 33%
	if overview.ApprovalRate != 33 {
		t.Fatalf("通过率期望 33%%，得到 %d", overview.ApprovalRate)
	}

	byTpl := svc.StatsByTemplate()
	if len(byTpl) != 1 || byTpl[0].RequestCount != 4 || byTpl[0].ApprovedCount != 1 {
		t.Fatalf("按模板统计不对: %+v", byTpl)
	}
	byDept := svc.StatsByDepartment()
	if len(byDept) != 1 || byDept[0].Department != "技术部" || byDept[0].RequestCount != 4 {
		t.Fatalf("按部门统计不对: %+v", byDept)
	}
	top := svc.TopApprovers(10)
	if len(top) != 2 {
		t.Fatalf("期望 2 个审批人，得到 %d", len(top))
	}
	// approver1 审批了 2 次（1 approve + 1 reject），排第一
	if top[0].OperatorID != approver1.ID || top[0].ApproveCount != 1 || top[0].RejectCount != 1 {
		t.Fatalf("审批人排行不对: %+v", top)
	}
}

func TestRequestListFilterAndPagination(t *testing.T) {
	svc := newTestService()
	tpl, applicant, _, _ := setupTwoLevelTemplate(t, svc)
	for i := 0; i < 5; i++ {
		if _, err := svc.SubmitRequest(tpl.ID, applicant.ID, "标题"+itoa(i), "事由", 0); err != nil {
			t.Fatal(err)
		}
	}
	items, total, err := svc.ListRequests(model.RequestFilter{Status: model.RequestPending}, 1, 10)
	if err != nil || total != 5 || len(items) != 5 {
		t.Fatalf("状态筛选失败: total=%d err=%v", total, err)
	}
	items, total, err = svc.ListRequests(model.RequestFilter{Keyword: "标题2"}, 1, 10)
	if err != nil || total != 1 {
		t.Fatalf("关键词筛选失败: total=%d err=%v", total, err)
	}
	// 分页
	items, total, err = svc.ListRequests(model.RequestFilter{}, 2, 3)
	if err != nil || total != 5 || len(items) != 2 {
		t.Fatalf("分页期望 total=5 len=2，得到 total=%d len=%d", total, len(items))
	}
}

func TestRecordListFilter(t *testing.T) {
	svc := newTestService()
	tpl, applicant, approver1, approver2 := setupTwoLevelTemplate(t, svc)
	req, _ := svc.SubmitRequest(tpl.ID, applicant.ID, "记录筛选", "x", 0)
	svc.Approve(req.ID, approver1.ID, "ok")
	svc.Approve(req.ID, approver2.ID, "ok")

	items, total, err := svc.ListRecords(model.RecordFilter{RequestID: req.ID}, 1, 10)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("按审批单筛选失败: total=%d err=%v", total, err)
	}
	_, total, err = svc.ListRecords(model.RecordFilter{OperatorID: approver1.ID}, 1, 10)
	if err != nil || total != 1 {
		t.Fatalf("按操作人筛选失败: total=%d err=%v", total, err)
	}
	_, total, err = svc.ListRecords(model.RecordFilter{Action: model.ActionReject}, 1, 10)
	if err != nil || total != 0 {
		t.Fatalf("无匹配应返回 0，得到 %d", total)
	}
	if _, err := svc.GetRecord(items[0].ID); err != nil {
		t.Fatal(err)
	}
}

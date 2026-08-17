package service

import (
	"sort"
	"time"

	"approvalflow/internal/model"
	"approvalflow/pkg/idgen"
)

// SubmitRequest 申请人基于模板发起审批。
// 校验链：模板存在 -> 模板 active -> 申请人存在且在职 -> 按模板节点生成审批流。
func (s *Service) SubmitRequest(templateID, applicantID, title, reason string, amount int64) (*model.ApprovalRequest, error) {
	t, err := s.store.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}
	if t.Status != model.TemplateActive {
		return nil, model.NewValidationError("template", "模板未生效，不能发起审批")
	}
	applicant, err := s.store.GetApplicant(applicantID)
	if err != nil {
		return nil, err
	}
	if applicant.Status != model.ApplicantActive {
		return nil, model.NewValidationError("applicant", "申请人已停用，不能发起审批")
	}
	nodes := s.nodesOfTemplate(templateID)
	if len(nodes) == 0 {
		return nil, model.NewValidationError("template", "模板未配置审批节点")
	}

	now := time.Now()
	req := &model.ApprovalRequest{
		ID:          idgen.Hex(),
		SerialNo:    "AP" + now.Format("20060102") + idgen.Short(),
		TemplateID:  templateID,
		ApplicantID: applicantID,
		Title:       title,
		Reason:      reason,
		Amount:      amount,
		CurrentSeq:  nodes[0].Seq,
		Status:      model.RequestPending,
		SubmittedAt: now,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.CreateRequest(req); err != nil {
		return nil, err
	}
	s.log.Infof("申请人 %s 提交审批单 %s（模板 %s）", applicant.Name, req.SerialNo, t.Code)
	return req, nil
}

// GetRequest 查询审批单详情。
func (s *Service) GetRequest(id string) (*model.ApprovalRequest, error) {
	return s.store.GetRequest(id)
}

// ListRequests 分页查询审批单列表，按提交时间倒序。
func (s *Service) ListRequests(filter model.RequestFilter, page, size int) ([]*model.ApprovalRequest, int, error) {
	all := s.store.ListRequests()
	matched := make([]*model.ApprovalRequest, 0, len(all))
	for _, r := range all {
		if filter.Match(r) {
			matched = append(matched, r)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].SubmittedAt.After(matched[j].SubmittedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.ApprovalRequest{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// PendingOf 查询某审批人的待办：所有 pending 且当前节点审批人是他的审批单。
func (s *Service) PendingOf(approverID string) ([]*model.ApprovalRequest, error) {
	if _, err := s.store.GetApplicant(approverID); err != nil {
		return nil, err
	}
	nodeByKey := make(map[string]*model.ApprovalNode)
	for _, n := range s.store.ListNodes() {
		nodeByKey[n.TemplateID+"#"+itoa(n.Seq)] = n
	}
	pending := make([]*model.ApprovalRequest, 0)
	for _, r := range s.store.ListRequests() {
		if r.Status != model.RequestPending {
			continue
		}
		n, ok := nodeByKey[r.TemplateID+"#"+itoa(r.CurrentSeq)]
		if ok && n.ApproverID == approverID {
			pending = append(pending, r)
		}
	}
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].SubmittedAt.Before(pending[j].SubmittedAt)
	})
	return pending, nil
}

// Approve 当前节点审批人同意。
// 若已是最后一个节点则审批单通过；否则 CurrentSeq 前进到下一节点。
func (s *Service) Approve(requestID, operatorID, comment string) (*model.ApprovalRequest, error) {
	req, node, err := s.checkOperable(requestID, operatorID)
	if err != nil {
		return nil, err
	}
	if err := s.writeRecord(requestID, node.Seq, operatorID, model.ActionApprove, comment); err != nil {
		return nil, err
	}
	nodes := s.nodesOfTemplate(req.TemplateID)
	if node.Seq == nodes[len(nodes)-1].Seq {
		req.Status = model.RequestApproved
		req.FinishedAt = time.Now()
		s.log.Infof("审批单 %s 全部节点通过，状态置为 approved", req.SerialNo)
	} else {
		req.CurrentSeq = nextSeq(nodes, node.Seq)
		s.log.Infof("审批单 %s 节点 #%d 通过，流转到节点 #%d", req.SerialNo, node.Seq, req.CurrentSeq)
	}
	if err := s.store.UpdateRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

// Reject 当前节点审批人驳回，审批单直接进入 rejected 终态。
func (s *Service) Reject(requestID, operatorID, comment string) (*model.ApprovalRequest, error) {
	req, node, err := s.checkOperable(requestID, operatorID)
	if err != nil {
		return nil, err
	}
	if comment == "" {
		return nil, model.NewValidationError("comment", "驳回必须填写审批意见")
	}
	if err := s.writeRecord(requestID, node.Seq, operatorID, model.ActionReject, comment); err != nil {
		return nil, err
	}
	req.Status = model.RequestRejected
	req.FinishedAt = time.Now()
	if err := s.store.UpdateRequest(req); err != nil {
		return nil, err
	}
	s.log.Infof("审批单 %s 被节点 #%d 驳回", req.SerialNo, node.Seq)
	return req, nil
}

// Cancel 申请人撤回自己 pending 的审批单。
func (s *Service) Cancel(requestID, applicantID, comment string) (*model.ApprovalRequest, error) {
	req, err := s.store.GetRequest(requestID)
	if err != nil {
		return nil, err
	}
	if req.ApplicantID != applicantID {
		return nil, model.NewValidationError("operator", "仅申请人本人可撤回")
	}
	if !model.CanRequestTransition(req.Status, model.RequestCanceled) {
		return nil, model.NewValidationError("status", "当前状态不可撤回")
	}
	if err := s.writeRecord(requestID, 0, applicantID, model.ActionCancel, comment); err != nil {
		return nil, err
	}
	req.Status = model.RequestCanceled
	req.FinishedAt = time.Now()
	if err := s.store.UpdateRequest(req); err != nil {
		return nil, err
	}
	s.log.Infof("审批单 %s 被申请人撤回", req.SerialNo)
	return req, nil
}

// checkOperable 校验审批单可被当前操作人审批：存在、pending、操作人是当前节点审批人。
func (s *Service) checkOperable(requestID, operatorID string) (*model.ApprovalRequest, *model.ApprovalNode, error) {
	req, err := s.store.GetRequest(requestID)
	if err != nil {
		return nil, nil, err
	}
	if req.Status != model.RequestPending {
		return nil, nil, model.NewValidationError("status", "审批单已结束，不可再审批")
	}
	var current *model.ApprovalNode
	for _, n := range s.nodesOfTemplate(req.TemplateID) {
		if n.Seq == req.CurrentSeq {
			current = n
			break
		}
	}
	if current == nil {
		return nil, nil, model.NewValidationError("node", "当前审批节点不存在")
	}
	if current.ApproverID != operatorID {
		return nil, nil, model.NewValidationError("operator", "当前节点审批人不是该操作人")
	}
	return req, current, nil
}

// writeRecord 写入一条审批记录。
func (s *Service) writeRecord(requestID string, nodeSeq int, operatorID, action, comment string) error {
	record := &model.ApprovalRecord{
		ID:         idgen.Hex(),
		RequestID:  requestID,
		NodeSeq:    nodeSeq,
		OperatorID: operatorID,
		Action:     action,
		Comment:    comment,
		CreatedAt:  time.Now(),
	}
	if err := record.Validate(); err != nil {
		return err
	}
	return s.store.CreateRecord(record)
}

// nextSeq 返回节点列表中大于 seq 的最小 Seq。
func nextSeq(nodes []*model.ApprovalNode, seq int) int {
	next := 0
	for _, n := range nodes {
		if n.Seq < seq && (next == 0 || n.Seq > next) {
			next = n.Seq
		}
	}
	return next
}

// itoa 整数转字符串（避免引入 strconv 的轻量实现）。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	buf := make([]byte, 0, 8)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	if negative {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

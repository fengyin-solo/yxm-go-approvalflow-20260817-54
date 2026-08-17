// Package store 定义数据访问接口与内存实现。
package store

import (
	"errors"

	"approvalflow/internal/model"
)

var (
	ErrNotFound = errors.New("记录不存在")
	ErrConflict = errors.New("记录已存在或状态冲突")
)

// Store 聚合全部实体的数据访问方法，便于测试时替换实现。
type Store interface {
	// 审批模板
	CreateTemplate(t *model.ApprovalTemplate) error
	GetTemplate(id string) (*model.ApprovalTemplate, error)
	GetTemplateByCode(code string) (*model.ApprovalTemplate, error)
	ListTemplates() []*model.ApprovalTemplate
	UpdateTemplate(t *model.ApprovalTemplate) error
	DeleteTemplate(id string) error

	// 申请人
	CreateApplicant(a *model.Applicant) error
	GetApplicant(id string) (*model.Applicant, error)
	GetApplicantByEmployeeNo(no string) (*model.Applicant, error)
	ListApplicants() []*model.Applicant
	UpdateApplicant(a *model.Applicant) error
	DeleteApplicant(id string) error

	// 审批节点
	CreateNode(n *model.ApprovalNode) error
	GetNode(id string) (*model.ApprovalNode, error)
	ListNodes() []*model.ApprovalNode
	DeleteNode(id string) error
	DeleteNodesByTemplate(templateID string) int

	// 审批单
	CreateRequest(r *model.ApprovalRequest) error
	GetRequest(id string) (*model.ApprovalRequest, error)
	GetRequestBySerialNo(serialNo string) (*model.ApprovalRequest, error)
	ListRequests() []*model.ApprovalRequest
	UpdateRequest(r *model.ApprovalRequest) error

	// 审批记录
	CreateRecord(r *model.ApprovalRecord) error
	GetRecord(id string) (*model.ApprovalRecord, error)
	ListRecords() []*model.ApprovalRecord
}

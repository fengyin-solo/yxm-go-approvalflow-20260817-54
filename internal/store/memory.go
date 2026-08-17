package store

import (
	"sync"

	"approvalflow/internal/model"
)

// MemoryStore 基于内存 map 的 Store 实现，所有操作线程安全。
type MemoryStore struct {
	mu         sync.RWMutex
	templates  map[string]*model.ApprovalTemplate
	applicants map[string]*model.Applicant
	nodes      map[string]*model.ApprovalNode
	requests   map[string]*model.ApprovalRequest
	records    map[string]*model.ApprovalRecord
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		templates:  make(map[string]*model.ApprovalTemplate),
		applicants: make(map[string]*model.Applicant),
		nodes:      make(map[string]*model.ApprovalNode),
		requests:   make(map[string]*model.ApprovalRequest),
		records:    make(map[string]*model.ApprovalRecord),
	}
}

var _ Store = (*MemoryStore)(nil)

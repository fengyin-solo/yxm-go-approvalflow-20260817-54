package store

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"approvalflow/internal/model"
)

func TestConcurrentMemoryStoreAccessIsRaceFree(t *testing.T) {
	s := NewMemoryStore()
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.CreateApplicant(&model.Applicant{ID: fmt.Sprintf("a-%d", i), EmployeeNo: fmt.Sprintf("E%03d", i), Name: "员工", Department: "研发", Status: model.ApplicantActive, CreatedAt: time.Now()})
			_ = s.CreateTemplate(&model.ApprovalTemplate{ID: fmt.Sprintf("t-%d", i), Code: fmt.Sprintf("T%03d", i), Name: "模板", Category: model.CategoryGeneral, Status: model.TemplateDraft, CreatedAt: time.Now()})
			_ = s.CreateRequest(&model.ApprovalRequest{ID: fmt.Sprintf("r-%d", i), SerialNo: fmt.Sprintf("AP%03d", i), TemplateID: "t", ApplicantID: "a", Title: "审批", Reason: "事由", CurrentSeq: 1, Status: model.RequestPending, SubmittedAt: time.Now()})
			_ = s.CreateRecord(&model.ApprovalRecord{ID: fmt.Sprintf("rec-%d", i), RequestID: "r", NodeSeq: 1, OperatorID: "a", Action: model.ActionApprove, CreatedAt: time.Now()})
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.ListApplicants(); _ = s.ListTemplates(); _ = s.ListRequests(); _ = s.ListRecords()
		}()
	}
	wg.Wait()
}

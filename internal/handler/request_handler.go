package handler

import (
	"net/http"

	"approvalflow/internal/model"
	"approvalflow/pkg/httpx"
)

func (s *Server) registerRequestRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/requests", s.submitRequest)
	mux.HandleFunc("GET /api/requests", s.listRequests)
	mux.HandleFunc("GET /api/requests/{id}", s.getRequest)
	mux.HandleFunc("GET /api/requests/{id}/timeline", s.requestTimeline)
	mux.HandleFunc("POST /api/requests/{id}/approve", s.approveRequest)
	mux.HandleFunc("POST /api/requests/{id}/reject", s.rejectRequest)
	mux.HandleFunc("POST /api/requests/{id}/cancel", s.cancelRequest)
	mux.HandleFunc("GET /api/applicants/{id}/pending", s.pendingOf)
}

type submitRequestPayload struct {
	TemplateID  string `json:"template_id"`
	ApplicantID string `json:"applicant_id"`
	Title       string `json:"title"`
	Reason      string `json:"reason"`
	Amount      int64  `json:"amount"`
}

func (s *Server) submitRequest(w http.ResponseWriter, r *http.Request) {
	var req submitRequestPayload
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	request, err := s.svc.SubmitRequest(req.TemplateID, req.ApplicantID, req.Title, req.Reason, req.Amount)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, request)
}

func (s *Server) listRequests(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.RequestFilter{
		TemplateID:  r.URL.Query().Get("template_id"),
		ApplicantID: r.URL.Query().Get("applicant_id"),
		Status:      r.URL.Query().Get("status"),
		Keyword:     r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListRequests(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getRequest(w http.ResponseWriter, r *http.Request) {
	request, err := s.svc.GetRequest(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, request)
}

func (s *Server) requestTimeline(w http.ResponseWriter, r *http.Request) {
	records, err := s.svc.Timeline(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, records)
}

type approveRequestPayload struct {
	OperatorID string `json:"operator_id"`
	Comment    string `json:"comment"`
}

func (s *Server) approveRequest(w http.ResponseWriter, r *http.Request) {
	var req approveRequestPayload
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	request, err := s.svc.Approve(r.PathValue("id"), req.OperatorID, req.Comment)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, request)
}

type rejectRequestPayload struct {
	OperatorID string `json:"operator_id"`
	Comment    string `json:"comment"`
}

func (s *Server) rejectRequest(w http.ResponseWriter, r *http.Request) {
	var req rejectRequestPayload
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	request, err := s.svc.Reject(r.PathValue("id"), req.OperatorID, req.Comment)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, request)
}

type cancelRequestPayload struct {
	ApplicantID string `json:"applicant_id"`
	Comment     string `json:"comment"`
}

func (s *Server) cancelRequest(w http.ResponseWriter, r *http.Request) {
	var req cancelRequestPayload
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	request, err := s.svc.Cancel(r.PathValue("id"), req.ApplicantID, req.Comment)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, request)
}

func (s *Server) pendingOf(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.PendingOf(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

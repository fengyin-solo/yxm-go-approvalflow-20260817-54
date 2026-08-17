package handler

import (
	"net/http"

	"approvalflow/internal/model"
	"approvalflow/pkg/httpx"
)

func (s *Server) registerApplicantRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/applicants", s.createApplicant)
	mux.HandleFunc("GET /api/applicants", s.listApplicants)
	mux.HandleFunc("GET /api/applicants/{id}", s.getApplicant)
	mux.HandleFunc("PUT /api/applicants/{id}", s.updateApplicant)
	mux.HandleFunc("DELETE /api/applicants/{id}", s.deleteApplicant)
}

type createApplicantRequest struct {
	EmployeeNo string `json:"employee_no"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Email      string `json:"email"`
}

func (s *Server) createApplicant(w http.ResponseWriter, r *http.Request) {
	var req createApplicantRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.CreateApplicant(model.Applicant{
		EmployeeNo: req.EmployeeNo,
		Name:       req.Name,
		Department: req.Department,
		Email:      req.Email,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, a)
}

func (s *Server) listApplicants(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ApplicantFilter{
		Department: r.URL.Query().Get("department"),
		Status:     r.URL.Query().Get("status"),
		Keyword:    r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListApplicants(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getApplicant(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.GetApplicant(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

type updateApplicantRequest struct {
	Name       string `json:"name"`
	Department string `json:"department"`
	Email      string `json:"email"`
	Status     string `json:"status"`
}

func (s *Server) updateApplicant(w http.ResponseWriter, r *http.Request) {
	var req updateApplicantRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.UpdateApplicant(r.PathValue("id"), req.Name, req.Department, req.Email, req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) deleteApplicant(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteApplicant(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

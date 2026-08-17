package handler

import (
	"net/http"

	"approvalflow/internal/model"
	"approvalflow/pkg/httpx"
)

func (s *Server) registerTemplateRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/templates", s.createTemplate)
	mux.HandleFunc("GET /api/templates", s.listTemplates)
	mux.HandleFunc("GET /api/templates/{id}", s.getTemplate)
	mux.HandleFunc("PUT /api/templates/{id}", s.updateTemplate)
	mux.HandleFunc("DELETE /api/templates/{id}", s.deleteTemplate)
	mux.HandleFunc("POST /api/templates/{id}/transition", s.transitionTemplate)
}

type createTemplateRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

func (s *Server) createTemplate(w http.ResponseWriter, r *http.Request) {
	var req createTemplateRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	tpl, err := s.svc.CreateTemplate(model.ApprovalTemplate{
		Code:        req.Code,
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, tpl)
}

func (s *Server) listTemplates(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.TemplateFilter{
		Status:   r.URL.Query().Get("status"),
		Category: r.URL.Query().Get("category"),
		Keyword:  r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListTemplates(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getTemplate(w http.ResponseWriter, r *http.Request) {
	tpl, err := s.svc.GetTemplate(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, tpl)
}

type updateTemplateRequest struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

func (s *Server) updateTemplate(w http.ResponseWriter, r *http.Request) {
	var req updateTemplateRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	tpl, err := s.svc.UpdateTemplate(r.PathValue("id"), req.Name, req.Category, req.Description)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, tpl)
}

func (s *Server) deleteTemplate(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteTemplate(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type transitionRequest struct {
	To string `json:"to"`
}

func (s *Server) transitionTemplate(w http.ResponseWriter, r *http.Request) {
	var req transitionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	tpl, err := s.svc.TransitionTemplate(r.PathValue("id"), req.To)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, tpl)
}

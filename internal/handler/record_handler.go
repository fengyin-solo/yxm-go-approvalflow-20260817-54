package handler

import (
	"net/http"

	"approvalflow/internal/model"
	"approvalflow/pkg/httpx"
)

func (s *Server) registerRecordRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/records", s.listRecords)
	mux.HandleFunc("GET /api/records/{id}", s.getRecord)
}

func (s *Server) listRecords(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.RecordFilter{
		RequestID:  r.URL.Query().Get("request_id"),
		OperatorID: r.URL.Query().Get("operator_id"),
		Action:     r.URL.Query().Get("action"),
	}
	items, total, err := s.svc.ListRecords(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getRecord(w http.ResponseWriter, r *http.Request) {
	record, err := s.svc.GetRecord(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, record)
}

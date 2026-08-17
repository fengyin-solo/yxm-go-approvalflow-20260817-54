package handler

import (
	"net/http"
	"strconv"

	"approvalflow/pkg/httpx"
)

func (s *Server) registerStatsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/stats/overview", s.statsOverview)
	mux.HandleFunc("GET /api/stats/by-template", s.statsByTemplate)
	mux.HandleFunc("GET /api/stats/by-department", s.statsByDepartment)
	mux.HandleFunc("GET /api/stats/top-approvers", s.statsTopApprovers)
}

func (s *Server) statsOverview(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.Stats())
}

func (s *Server) statsByTemplate(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.StatsByTemplate())
}

func (s *Server) statsByDepartment(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.StatsByDepartment())
}

func (s *Server) statsTopApprovers(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.URL.Query().Get("n"))
	httpx.OK(w, s.svc.TopApprovers(n))
}

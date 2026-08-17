package handler

import (
	"net/http"

	"approvalflow/pkg/httpx"
)

func (s *Server) registerNodeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/templates/{id}/nodes", s.addNode)
	mux.HandleFunc("GET /api/templates/{id}/nodes", s.listNodes)
	mux.HandleFunc("DELETE /api/nodes/{id}", s.removeNode)
}

type addNodeRequest struct {
	Seq        int    `json:"seq"`
	Name       string `json:"name"`
	ApproverID string `json:"approver_id"`
}

func (s *Server) addNode(w http.ResponseWriter, r *http.Request) {
	var req addNodeRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	node, err := s.svc.AddNode(r.PathValue("id"), req.Seq, req.Name, req.ApproverID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, node)
}

func (s *Server) listNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.svc.ListNodes(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, nodes)
}

func (s *Server) removeNode(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.RemoveNode(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

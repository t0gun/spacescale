// Http api handlers for app and deployment routes
// Handlers decode json and call the service
// Errors are mapped to http status codes
// Success responses write json payloads
// This file ties routing to service logic

package http_api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/t0gun/spacescale/internal/service"
)

// This function handles handle create app
// It supports handle create app behavior
func (s *Server) handleCreateApp(w http.ResponseWriter, r *http.Request) {
	var req createAppReq
	if err := readJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
	}

	app, err := s.svc.CreateApp(r.Context(), service.CreateAppParams{
		Name:   req.Name,
		Image:  req.Image,
		Port:   req.Port,
		Expose: req.Expose,
		Env:    req.Env,
	})
	if err != nil {
		status, msg := mapServiceErr(err)
		if status == http.StatusNoContent {
			w.WriteHeader(status)
			return
		}
		writeErr(w, status, msg)
		return
	}
	writeJSON(w, http.StatusCreated, toAppResp(app))
}

// This function handles handle deploy app
// It supports handle deploy app behavior
func (s *Server) handleDeployApp(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appID")
	dep, err := s.svc.DeployApp(r.Context(), service.DeployAppParams{AppID: appID})
	if err != nil {
		status, msg := mapServiceErr(err)
		if status == http.StatusNoContent {
			w.WriteHeader(status)
			return
		}
		writeErr(w, status, msg)
		return
	}
	writeJSON(w, http.StatusAccepted, toDeploymentResp(dep))
}

// This function handles handle list deployments
// It supports handle list deployments behavior
func (s *Server) handleListDeployments(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appID")
	deps, err := s.svc.ListDeployments(r.Context(), service.ListDeploymentsParams{AppID: appID})
	if err != nil {
		status, msg := mapServiceErr(err)
		if status == http.StatusNoContent {
			w.WriteHeader(status)
			return
		}
		writeErr(w, status, msg)
		return
	}
	out := make([]deploymentResp, 0, len(deps))
	for _, d := range deps {
		out = append(out, toDeploymentResp(d))
	}
	writeJSON(w, http.StatusOK, out)
}

// This function handles handle process next deployment
// It supports handle process next deployment behavior
func (s *Server) handleProcessNextDeployment(w http.ResponseWriter, r *http.Request) {
	dep, err := s.svc.ProcessNextDeployment(r.Context())
	if err != nil {
		status, msg := mapServiceErr(err)
		if status == http.StatusNoContent {
			w.WriteHeader(status)
			return
		}
		writeErr(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, toDeploymentResp(dep))
}

// This function handles handle list apps
// It supports handle list apps behavior
func (s *Server) handleListApps(w http.ResponseWriter, r *http.Request) {
	apps, err := s.svc.ListApps(r.Context())
	if err != nil {
		status, msg := mapServiceErr(err)
		if status == http.StatusNoContent {
			w.WriteHeader(status)
			return
		}
		writeErr(w, status, msg)
		return
	}
	out := make([]appResp, 0, len(apps))
	for _, a := range apps {
		out = append(out, toAppResp(a))
	}
	writeJSON(w, http.StatusOK, out)
}

// This function handles handle get app by id
// It supports handle get app by id behavior
func (s *Server) handleGetAppByID(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appID")
	app, err := s.svc.GetAppByID(r.Context(), appID)
	if err != nil {
		status, msg := mapServiceErr(err)
		if status == http.StatusNoContent {
			w.WriteHeader(status)
			return
		}
		writeErr(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, toAppResp(app))
}

package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/metaclaw/metaclaw-registry/internal/model"
	"github.com/metaclaw/metaclaw-registry/internal/service"
)

type Handler struct {
	svc        *service.Service
	adminToken string
}

func New(svc *service.Service, adminToken string) *Handler {
	return &Handler{svc: svc, adminToken: strings.TrimSpace(adminToken)}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET /v1/artifacts", h.listArtifacts)
	mux.HandleFunc("POST /v1/artifacts", h.createArtifact)
	mux.HandleFunc("GET /v1/artifacts/", h.getArtifact)
	return mux
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listArtifacts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 100
	if raw := strings.TrimSpace(q.Get("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	filter := model.ListFilter{
		Kind:  model.ArtifactKind(strings.TrimSpace(q.Get("kind"))),
		Name:  strings.TrimSpace(q.Get("name")),
		Limit: limit,
	}
	items := h.svc.List(filter)
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

func (h *Handler) createArtifact(w http.ResponseWriter, r *http.Request) {
	if err := h.authorize(r); err != nil {
		writeErr(w, http.StatusUnauthorized, err)
		return
	}
	var in model.Artifact
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	created, err := h.svc.Register(in)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) getArtifact(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/artifacts/")
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		writeErr(w, http.StatusBadRequest, errors.New("path must be /v1/artifacts/{kind}/{name}/{version}"))
		return
	}
	kind := model.ArtifactKind(parts[0])
	name := parts[1]
	version := parts[2]
	a, ok := h.svc.Get(kind, name, version)
	if !ok {
		writeErr(w, http.StatusNotFound, errors.New("artifact not found"))
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) authorize(r *http.Request) error {
	if h.adminToken == "" {
		return nil
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	want := "Bearer " + h.adminToken
	if auth != want {
		return errors.New("missing or invalid bearer token")
	}
	return nil
}

func writeErr(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]any{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/fpp-125/metaclaw-registry/internal/model"
	"github.com/fpp-125/metaclaw-registry/internal/service"
	"github.com/fpp-125/metaclaw-registry/internal/store"
)

func newTestHandler(t *testing.T, token string) http.Handler {
	t.Helper()
	st, err := store.NewFileStore(filepath.Join(t.TempDir(), "registry.json"))
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}
	svc := service.New(st)
	return New(svc, token).Routes()
}

func TestHealthz(t *testing.T) {
	h := newTestHandler(t, "")
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("healthz code=%d", rr.Code)
	}
}

func TestCreateAndGetArtifact(t *testing.T) {
	h := newTestHandler(t, "token123")
	body := model.Artifact{
		Kind:    model.KindSkill,
		Name:    "obsidian.search",
		Version: "v1.0.0",
		Digest:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		OCIRef:  "ghcr.io/metaclaw/skills/obsidian.search:v1.0.0",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/artifacts", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer token123")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create code=%d body=%s", rr.Code, rr.Body.String())
	}

	reqGet := httptest.NewRequest(http.MethodGet, "/v1/artifacts/skill/obsidian.search/v1.0.0", nil)
	rrGet := httptest.NewRecorder()
	h.ServeHTTP(rrGet, reqGet)
	if rrGet.Code != http.StatusOK {
		t.Fatalf("get code=%d body=%s", rrGet.Code, rrGet.Body.String())
	}
}

func TestCreateArtifactRejectsWithoutAuth(t *testing.T) {
	h := newTestHandler(t, "token123")
	body := model.Artifact{
		Kind:    model.KindSkill,
		Name:    "obsidian.search",
		Version: "v1.0.0",
		Digest:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		OCIRef:  "ghcr.io/metaclaw/skills/obsidian.search:v1.0.0",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/artifacts", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestListArtifacts(t *testing.T) {
	h := newTestHandler(t, "")
	items := []model.Artifact{
		{
			Kind:    model.KindSkill,
			Name:    "obsidian.search",
			Version: "v1.0.0",
			Digest:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			OCIRef:  "ghcr.io/metaclaw/skills/obsidian.search:v1.0.0",
		},
		{
			Kind:    model.KindCapsule,
			Name:    "obsidian-bot",
			Version: "v0.1.0",
			Digest:  "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			OCIRef:  "ghcr.io/metaclaw/capsules/obsidian-bot:v0.1.0",
		},
	}
	for _, item := range items {
		b, _ := json.Marshal(item)
		req := httptest.NewRequest(http.MethodPost, "/v1/artifacts", bytes.NewReader(b))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("create code=%d", rr.Code)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/artifacts?kind=skill", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("list code=%d", rr.Code)
	}
	var payload struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if payload.Count != 1 {
		t.Fatalf("expected 1 item, got %d", payload.Count)
	}
}

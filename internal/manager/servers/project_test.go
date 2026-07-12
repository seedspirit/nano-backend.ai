package servers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/seedspirit/nano-backend.ai/internal/manager/repository"
	"github.com/seedspirit/nano-backend.ai/internal/manager/service"
	"github.com/seedspirit/nano-backend.ai/internal/manager/service/projectsvc"
	"github.com/seedspirit/nano-backend.ai/internal/manager/service/sessionsvc"
)

func TestProjectCreateAndGetAPI(t *testing.T) {
	handler := newProjectTestHandler(t)
	created := requestJSON(t, handler, http.MethodPost, "/v1/projects", `{
		"name": "  mergeowl  ",
		"description": "model development"
	}`, http.StatusCreated)
	data := requireMap(t, created, "data")
	if data["name"] != "mergeowl" {
		t.Fatalf("created name = %v, want mergeowl", data["name"])
	}

	projectID := requireString(t, data, "id")
	got := requestJSON(t, handler, http.MethodGet, "/v1/projects/"+projectID, "", http.StatusOK)
	gotData := requireMap(t, got, "data")
	if gotData["id"] != projectID || gotData["description"] != "model development" {
		t.Fatalf("GET data = %v, want created project", gotData)
	}
}

func requireMap(t *testing.T, source map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := source[key].(map[string]any)
	if !ok {
		t.Fatalf("%s has type %T, want object", key, source[key])
	}
	return value
}

func requireString(t *testing.T, source map[string]any, key string) string {
	t.Helper()
	value, ok := source[key].(string)
	if !ok {
		t.Fatalf("%s has type %T, want string", key, source[key])
	}
	return value
}

func TestProjectCreateAPIRejectsInvalidAndDuplicateName(t *testing.T) {
	handler := newProjectTestHandler(t)
	invalid := requestJSON(t, handler, http.MethodPost, "/v1/projects", `{"name":"  "}`, http.StatusUnprocessableEntity)
	if code := requireMap(t, invalid, "error")["code"]; code != "validation_error" {
		t.Fatalf("invalid name code = %v, want validation_error", code)
	}

	requestJSON(t, handler, http.MethodPost, "/v1/projects", `{"name":"mergeowl"}`, http.StatusCreated)
	duplicate := requestJSON(t, handler, http.MethodPost, "/v1/projects", `{"name":"mergeowl"}`, http.StatusConflict)
	if code := requireMap(t, duplicate, "error")["code"]; code != "project_name_conflict" {
		t.Fatalf("duplicate name code = %v, want project_name_conflict", code)
	}
}

func TestProjectGetAPIRejectsInvalidID(t *testing.T) {
	handler := newProjectTestHandler(t)
	got := requestJSON(t, handler, http.MethodGet, "/v1/projects/not-a-uuid", "", http.StatusBadRequest)
	if code := requireMap(t, got, "error")["code"]; code != "invalid_project_id" {
		t.Fatalf("error code = %v, want invalid_project_id", code)
	}
}

func newProjectTestHandler(t *testing.T) http.Handler {
	t.Helper()
	repositories, err := repository.NewRepositories(context.Background(), repository.Args{
		DBPath: filepath.Join(t.TempDir(), "manager.db"),
	})
	if err != nil {
		t.Fatalf("NewRepositories: %v", err)
	}
	t.Cleanup(func() {
		if err := repositories.Close(); err != nil {
			t.Errorf("Close repositories: %v", err)
		}
	})

	services := service.NewServices().
		WithProjectService(projectsvc.Args{Repositories: repositories}).
		WithSessionService(sessionsvc.Args{Repositories: repositories})
	server, err := NewServer(ServerArgs{Services: services})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	return server.httpServer.Handler
}

func requestJSON(t *testing.T, handler http.Handler, method, path, body string, wantStatus int) map[string]any {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != wantStatus {
		t.Fatalf("%s %s status = %d, want %d; body=%s", method, path, recorder.Code, wantStatus, recorder.Body.String())
	}

	var result map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, recorder.Body.String())
	}
	return result
}

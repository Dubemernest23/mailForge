package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mailForgeApi/internal/shared/constants"
	"mailForgeApi/internal/shared/response"
	"mailForgeApi/pkg/logger"
)

func TestHealthCheckReturnsOK(t *testing.T) {
	router := NewRouter(logger.New("test"), nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != constants.StatusOK {
		t.Fatalf("expected status %d, got %d", constants.StatusOK, rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected health status ok, got %q", body["status"])
	}
}

func TestRouterReturnsUniformNotFoundError(t *testing.T) {
	router := NewRouter(logger.New("test"), nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, constants.StatusNotFound, response.CodeRouteNotFound)
}

func TestRouterReturnsUniformMethodNotAllowedError(t *testing.T) {
	router := NewRouter(logger.New("test"), nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, constants.StatusMethodNotAllowed, response.CodeMethodNotAllowed)
}

func TestSwaggerDocIsServedOutsideProduction(t *testing.T) {
	t.Setenv("APP_ENV", "development")

	router := NewRouter(logger.New("test"), nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != constants.StatusOK {
		t.Fatalf("expected status %d, got %d", constants.StatusOK, rec.Code)
	}

	var spec map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&spec); err != nil {
		t.Fatalf("failed to decode Swagger doc: %v", err)
	}
	if spec["openapi"] != "3.0.0" {
		t.Fatalf("expected OpenAPI version 3.0.0, got %v", spec["openapi"])
	}
}

func assertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int, expectedCode string) {
	t.Helper()

	if rec.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, rec.Code)
	}

	var body response.ErrorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}
	if body.Success {
		t.Fatal("expected error response success to be false")
	}
	if body.Error.Code != expectedCode {
		t.Fatalf("expected error code %q, got %q", expectedCode, body.Error.Code)
	}
	if body.Error.Message == "" {
		t.Fatal("expected error message to be populated")
	}
	if body.RequestID == "" {
		t.Fatal("expected request_id to be populated")
	}
}

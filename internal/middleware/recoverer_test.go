package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"mailForgeApi/internal/constants"
	"mailForgeApi/internal/response"
	"mailForgeApi/pkg/logger"
)

func TestRecovererWritesUniformInternalServerError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	handler := Recoverer(logger.New("test"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("unexpected failure")
	}))
	handler = chimiddleware.RequestID(handler)

	handler.ServeHTTP(rec, req)

	if rec.Code != constants.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", constants.StatusInternalServerError, rec.Code)
	}

	var body response.ErrorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}
	if body.Success {
		t.Fatal("expected success to be false")
	}
	if body.Error.Code != response.CodeInternalServerError {
		t.Fatalf("expected code %q, got %q", response.CodeInternalServerError, body.Error.Code)
	}
	if body.RequestID == "" {
		t.Fatal("expected request id to be populated")
	}
}

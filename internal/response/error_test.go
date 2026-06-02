package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"mailForgeApi/internal/constants"
)

func TestHandleErrorWritesAppErrorResponse(t *testing.T) {
	req := requestWithID(http.MethodGet, "/test")
	rec := httptest.NewRecorder()
	err := NewAppError(constants.StatusConflict, CodeConflict, "resource already exists")

	HandleError(rec, req, err)

	assertErrorBody(t, rec, constants.StatusConflict, CodeConflict, "resource already exists")
}

func TestHandleErrorWritesInternalServerErrorForUnknownError(t *testing.T) {
	req := requestWithID(http.MethodGet, "/test")
	rec := httptest.NewRecorder()

	HandleError(rec, req, errors.New("database exploded"))

	assertErrorBody(t, rec, constants.StatusInternalServerError, CodeInternalServerError, "an unexpected error occurred")
}

func TestHandlerConvertsReturnedErrorToUniformResponse(t *testing.T) {
	req := requestWithID(http.MethodPost, "/test")
	rec := httptest.NewRecorder()
	handler := Handler(func(w http.ResponseWriter, r *http.Request) error {
		return NewAppError(constants.StatusBadRequest, CodeBadRequest, "invalid payload")
	})

	handler.ServeHTTP(rec, req)

	assertErrorBody(t, rec, constants.StatusBadRequest, CodeBadRequest, "invalid payload")
}

func requestWithID(method string, target string) *http.Request {
	return httptest.NewRequest(method, target, nil)
}

func assertErrorBody(t *testing.T, rec *httptest.ResponseRecorder, statusCode int, code string, message string) {
	t.Helper()

	if rec.Code != statusCode {
		t.Fatalf("expected status %d, got %d", statusCode, rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != ContentTypeJSON {
		t.Fatalf("expected content type %q, got %q", ContentTypeJSON, contentType)
	}

	var body ErrorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}
	if body.Success {
		t.Fatal("expected success to be false")
	}
	if body.Error.Code != code {
		t.Fatalf("expected code %q, got %q", code, body.Error.Code)
	}
	if body.Error.Message != message {
		t.Fatalf("expected message %q, got %q", message, body.Error.Message)
	}
	if body.RequestID != "" {
		t.Fatalf("expected empty request id, got %q", body.RequestID)
	}
}

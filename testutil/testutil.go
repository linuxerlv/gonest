package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linuxerlv/gonest"
)

type TestResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

func NewTestContext(method, path string, body any) (*gonest.HttpContext, *httptest.ResponseRecorder) {
	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case []byte:
			bodyReader = bytes.NewReader(v)
		case string:
			bodyReader = bytes.NewReader([]byte(v))
		default:
			data, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(data)
		}
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()

	return gonest.NewContext(w, req), w
}

func NewTestContextWithHeaders(method, path string, body any, headers map[string]string) (*gonest.HttpContext, *httptest.ResponseRecorder) {
	ctx, w := NewTestContext(method, path, body)
	for k, v := range headers {
		ctx.Request().Header.Set(k, v)
	}
	return ctx, w
}

func ExecuteMiddleware(t *testing.T, ctx gonest.Context, middleware gonest.Middleware, handler gonest.RouteHandler) error {
	return middleware.Handle(ctx, func() error {
		return handler(ctx)
	})
}

func ParseJSONResponse(t *testing.T, w *httptest.ResponseRecorder, dest any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), dest); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
}

func AssertStatusCode(t *testing.T, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected status code %d, got %d", expected, actual)
	}
}

func AssertJSONEquals(t *testing.T, expected, actual string) {
	t.Helper()
	var expectedJSON, actualJSON any
	if err := json.Unmarshal([]byte(expected), &expectedJSON); err != nil {
		t.Fatalf("Invalid expected JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(actual), &actualJSON); err != nil {
		t.Fatalf("Invalid actual JSON: %v", err)
	}
	if !compareJSON(expectedJSON, actualJSON) {
		t.Errorf("JSON mismatch:\nExpected: %s\nActual: %s", expected, actual)
	}
}

func compareJSON(a, b any) bool {
	switch va := a.(type) {
	case map[string]any:
		vb, ok := b.(map[string]any)
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if !compareJSON(v, vb[k]) {
				return false
			}
		}
		return true
	case []any:
		vb, ok := b.([]any)
		if !ok || len(va) != len(vb) {
			return false
		}
		for i := range va {
			if !compareJSON(va[i], vb[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

type MockStore struct {
	data map[string]any
}

func NewMockStore() *MockStore {
	return &MockStore{data: make(map[string]any)}
}

func (s *MockStore) Set(key string, value any) {
	s.data[key] = value
}

func (s *MockStore) Get(key string) any {
	return s.data[key]
}

func (s *MockStore) Delete(key string) {
	delete(s.data, key)
}

func (s *MockStore) Clear() {
	s.data = make(map[string]any)
}

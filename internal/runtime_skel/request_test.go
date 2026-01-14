package runtime

import (
	"context"
	"strings"
	"testing"
)

func TestNewRequest(t *testing.T) {
	req := NewRequest("GET", "/api/users")

	if req.Method != "GET" {
		t.Errorf("expected method GET, got %q", req.Method)
	}
	if req.Path != "/api/users" {
		t.Errorf("expected path /api/users, got %q", req.Path)
	}
	if req.PathParams == nil {
		t.Error("expected PathParams map to be initialized")
	}
	if req.QueryParams == nil {
		t.Error("expected QueryParams map to be initialized")
	}
	if req.Headers == nil {
		t.Error("expected Headers map to be initialized")
	}
}

func TestRequest_SetPathParam(t *testing.T) {
	req := NewRequest("GET", "/users/{id}")
	req.SetPathParam("id", "123")

	if req.PathParams["id"] != "123" {
		t.Errorf("expected path param id=123, got %q", req.PathParams["id"])
	}
}

func TestRequest_SetQueryParam(t *testing.T) {
	req := NewRequest("GET", "/users")
	req.SetQueryParam("page", "1")
	req.SetQueryParam("limit", "10")

	if req.QueryParams["page"] != "1" {
		t.Errorf("expected query param page=1, got %q", req.QueryParams["page"])
	}
	if req.QueryParams["limit"] != "10" {
		t.Errorf("expected query param limit=10, got %q", req.QueryParams["limit"])
	}
}

func TestRequest_SetHeader(t *testing.T) {
	req := NewRequest("POST", "/users")
	req.SetHeader("Authorization", "Bearer token123")

	if req.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("expected header Authorization='Bearer token123', got %q", req.Headers["Authorization"])
	}
}

func TestRequest_SetBody(t *testing.T) {
	req := NewRequest("POST", "/users")
	body := []byte(`{"name": "test"}`)
	req.SetBody(body)

	if string(req.Body) != string(body) {
		t.Errorf("expected body %q, got %q", string(body), string(req.Body))
	}
}

func TestRequest_Build_SimpleURL(t *testing.T) {
	ctx := context.Background()
	req := NewRequest("GET", "/api/users")

	httpReq, err := req.Build(ctx, "https://api.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedURL := "https://api.example.com/api/users"
	if httpReq.URL.String() != expectedURL {
		t.Errorf("expected URL %q, got %q", expectedURL, httpReq.URL.String())
	}
	if httpReq.Method != "GET" {
		t.Errorf("expected method GET, got %q", httpReq.Method)
	}
}

func TestRequest_Build_WithPathParams(t *testing.T) {
	ctx := context.Background()
	req := NewRequest("GET", "/users/{id}/posts/{postId}")
	req.SetPathParam("id", "123")
	req.SetPathParam("postId", "456")

	httpReq, err := req.Build(ctx, "https://api.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedURL := "https://api.example.com/users/123/posts/456"
	if httpReq.URL.String() != expectedURL {
		t.Errorf("expected URL %q, got %q", expectedURL, httpReq.URL.String())
	}
}

func TestRequest_Build_WithQueryParams(t *testing.T) {
	ctx := context.Background()
	req := NewRequest("GET", "/users")
	req.SetQueryParam("page", "1")
	req.SetQueryParam("limit", "10")

	httpReq, err := req.Build(ctx, "https://api.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that query params are present
	if !strings.Contains(httpReq.URL.String(), "page=1") {
		t.Errorf("expected URL to contain page=1, got %q", httpReq.URL.String())
	}
	if !strings.Contains(httpReq.URL.String(), "limit=10") {
		t.Errorf("expected URL to contain limit=10, got %q", httpReq.URL.String())
	}
}

func TestRequest_Build_WithHeaders(t *testing.T) {
	ctx := context.Background()
	req := NewRequest("GET", "/users")
	req.SetHeader("Authorization", "Bearer token123")
	req.SetHeader("X-Request-Id", "abc123")

	httpReq, err := req.Build(ctx, "https://api.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if httpReq.Header.Get("Authorization") != "Bearer token123" {
		t.Errorf("expected Authorization header, got %q", httpReq.Header.Get("Authorization"))
	}
	if httpReq.Header.Get("X-Request-Id") != "abc123" {
		t.Errorf("expected X-Request-Id header, got %q", httpReq.Header.Get("X-Request-Id"))
	}
}

func TestRequest_Build_WithBody(t *testing.T) {
	ctx := context.Background()
	req := NewRequest("POST", "/users")
	body := []byte(`{"name": "test"}`)
	req.SetBody(body)

	httpReq, err := req.Build(ctx, "https://api.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check Content-Type header is set for JSON body
	if httpReq.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", httpReq.Header.Get("Content-Type"))
	}
}

func TestRequest_Build_PathParamEncoding(t *testing.T) {
	ctx := context.Background()
	req := NewRequest("GET", "/users/{id}")
	req.SetPathParam("id", "hello world") // Contains space

	httpReq, err := req.Build(ctx, "https://api.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// When we access URL.RawPath or the full URL string, space should be encoded
	// url.PathEscape encodes space as %20
	fullURL := httpReq.URL.String()
	if !strings.Contains(fullURL, "hello%20world") {
		t.Errorf("expected URL-encoded path param in URL string, got %q", fullURL)
	}
}

func TestRequest_Build_BaseURLTrailingSlash(t *testing.T) {
	ctx := context.Background()
	req := NewRequest("GET", "/users")

	httpReq, err := req.Build(ctx, "https://api.example.com/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not have double slashes
	if strings.Contains(httpReq.URL.String(), "//users") {
		t.Errorf("expected no double slashes, got %q", httpReq.URL.String())
	}
}

func TestRequest_Build_MissingPathParam(t *testing.T) {
	ctx := context.Background()
	req := NewRequest("GET", "/users/{id}/posts/{postId}")
	req.SetPathParam("id", "123")
	// Missing postId

	_, err := req.Build(ctx, "https://api.example.com")
	if err == nil {
		t.Fatal("expected error for missing path param")
	}

	if !strings.Contains(err.Error(), "missing required path parameter") {
		t.Errorf("expected 'missing required path parameter' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "postId") {
		t.Errorf("expected error to mention 'postId', got: %v", err)
	}
}

func TestRequest_Build_AllPathParamsMissing(t *testing.T) {
	ctx := context.Background()
	req := NewRequest("GET", "/users/{id}/posts/{postId}")
	// No path params set

	_, err := req.Build(ctx, "https://api.example.com")
	if err == nil {
		t.Fatal("expected error for missing path params")
	}

	if !strings.Contains(err.Error(), "id") || !strings.Contains(err.Error(), "postId") {
		t.Errorf("expected error to mention both missing params, got: %v", err)
	}
}

func TestRequest_Build_NoPathParams(t *testing.T) {
	ctx := context.Background()
	req := NewRequest("GET", "/users")
	// No path params needed

	httpReq, err := req.Build(ctx, "https://api.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedURL := "https://api.example.com/users"
	if httpReq.URL.String() != expectedURL {
		t.Errorf("expected URL %q, got %q", expectedURL, httpReq.URL.String())
	}
}

package runtime

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
)

func TestIsJSON_ValidJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"object", `{"key": "value"}`, true},
		{"array", `[1, 2, 3]`, true},
		{"string", `"hello"`, true},
		{"number", `123`, true},
		{"boolean", `true`, true},
		{"null", `null`, true},
		{"nested", `{"a": {"b": [1, 2, 3]}}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isJSON([]byte(tt.input))
			if got != tt.want {
				t.Errorf("isJSON(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsJSON_InvalidJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"plain text", "hello world"},
		{"invalid json", "{invalid}"},
		{"trailing comma", `{"a": 1,}`},
		{"xml", "<root><child/></root>"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isJSON([]byte(tt.input))
			if got {
				t.Errorf("isJSON(%q) = true, want false", tt.input)
			}
		})
	}
}

func TestPrettyPrint_ValidJSON(t *testing.T) {
	input := `{"name":"test","value":123}`
	buf := new(bytes.Buffer)

	prettyPrint([]byte(input), buf)

	output := buf.String()
	// Should contain indentation
	if !strings.Contains(output, "  ") {
		t.Errorf("expected indented output, got %q", output)
	}
	// Should contain the values
	if !strings.Contains(output, `"name"`) {
		t.Errorf("expected output to contain 'name', got %q", output)
	}
	if !strings.Contains(output, `"test"`) {
		t.Errorf("expected output to contain 'test', got %q", output)
	}
}

func TestPrettyPrint_InvalidJSON(t *testing.T) {
	input := "not valid json"
	buf := new(bytes.Buffer)

	prettyPrint([]byte(input), buf)

	output := strings.TrimSpace(buf.String())
	// Should fall back to raw output
	if output != input {
		t.Errorf("expected raw output %q, got %q", input, output)
	}
}

func TestPrettyPrint_ComplexJSON(t *testing.T) {
	input := `{"array":[1,2,3],"nested":{"a":"b"},"bool":true,"null":null}`
	buf := new(bytes.Buffer)

	prettyPrint([]byte(input), buf)

	output := buf.String()
	// Should be properly formatted with newlines and indentation
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 5 {
		t.Errorf("expected multiple lines in pretty output, got %d lines", len(lines))
	}
}

type mockResponseBody struct {
	*bytes.Reader
}

func (m *mockResponseBody) Close() error {
	return nil
}

func TestHandleResponse_Success(t *testing.T) {
	body := []byte(`{"status": "ok"}`)
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Body:       &mockResponseBody{bytes.NewReader(body)},
	}

	buf := new(bytes.Buffer)
	err := handleResponse(resp, buf)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "status") {
		t.Errorf("expected output to contain response body, got %q", output)
	}
}

func TestHandleResponse_Error(t *testing.T) {
	body := []byte(`{"error": "not found"}`)
	resp := &http.Response{
		StatusCode: 404,
		Status:     "404 Not Found",
		Body:       &mockResponseBody{bytes.NewReader(body)},
	}

	buf := new(bytes.Buffer)
	err := handleResponse(resp, buf)

	if err == nil {
		t.Fatal("expected error for 404 response")
	}

	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected error to mention status code, got: %v", err)
	}
}

func TestHandleResponse_EmptyBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: 204,
		Status:     "204 No Content",
		Body:       &mockResponseBody{bytes.NewReader([]byte{})},
	}

	buf := new(bytes.Buffer)
	err := handleResponse(resp, buf)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Output should be empty for empty body
	if buf.Len() > 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestHandleResponse_NonJSONBody(t *testing.T) {
	body := []byte("plain text response")
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Body:       &mockResponseBody{bytes.NewReader(body)},
	}

	buf := new(bytes.Buffer)
	err := handleResponse(resp, buf)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != string(body) {
		t.Errorf("expected %q, got %q", string(body), output)
	}
}

func TestHandleResponse_ServerError(t *testing.T) {
	body := []byte(`{"error": "internal server error"}`)
	resp := &http.Response{
		StatusCode: 500,
		Status:     "500 Internal Server Error",
		Body:       &mockResponseBody{bytes.NewReader(body)},
	}

	buf := new(bytes.Buffer)
	err := handleResponse(resp, buf)

	if err == nil {
		t.Fatal("expected error for 500 response")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to mention status code, got: %v", err)
	}
}

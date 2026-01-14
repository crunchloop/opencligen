package runtime

import (
	"bytes"
	"strings"
	"testing"
)

func TestHandleSSE_BasicEvents(t *testing.T) {
	input := `data: {"message": "hello"}

data: {"message": "world"}

`

	reader := strings.NewReader(input)
	var out bytes.Buffer

	err := handleSSE(reader, &out)
	if err != nil {
		t.Fatalf("handleSSE failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, `"message": "hello"`) {
		t.Errorf("expected output to contain hello message, got: %s", output)
	}
	if !strings.Contains(output, `"message": "world"`) {
		t.Errorf("expected output to contain world message, got: %s", output)
	}
}

func TestHandleSSE_MultilineData(t *testing.T) {
	input := `data: {"line": 1,
data:  "continued": true}

`

	reader := strings.NewReader(input)
	var out bytes.Buffer

	err := handleSSE(reader, &out)
	if err != nil {
		t.Fatalf("handleSSE failed: %v", err)
	}

	output := out.String()
	// The multiline data should be combined
	if !strings.Contains(output, "line") {
		t.Errorf("expected output to contain 'line', got: %s", output)
	}
}

func TestHandleSSE_KeepAlives(t *testing.T) {
	input := `: keep-alive
data: {"status": "ok"}

: another keep-alive
data: {"status": "done"}

`

	reader := strings.NewReader(input)
	var out bytes.Buffer

	err := handleSSE(reader, &out)
	if err != nil {
		t.Fatalf("handleSSE failed: %v", err)
	}

	output := out.String()
	// Should not contain keep-alive comments
	if strings.Contains(output, "keep-alive") {
		t.Errorf("output should not contain keep-alive comments: %s", output)
	}
	// Should contain the data
	if !strings.Contains(output, "ok") {
		t.Errorf("expected output to contain 'ok', got: %s", output)
	}
	if !strings.Contains(output, "done") {
		t.Errorf("expected output to contain 'done', got: %s", output)
	}
}

func TestHandleSSE_EventAndIdFields(t *testing.T) {
	input := `event: message
id: 1
data: {"type": "test"}

`

	reader := strings.NewReader(input)
	var out bytes.Buffer

	err := handleSSE(reader, &out)
	if err != nil {
		t.Fatalf("handleSSE failed: %v", err)
	}

	output := out.String()
	// Should contain the data, but not event/id fields
	if !strings.Contains(output, "test") {
		t.Errorf("expected output to contain 'test', got: %s", output)
	}
}

func TestHandleSSE_PlainTextData(t *testing.T) {
	input := `data: This is plain text

`

	reader := strings.NewReader(input)
	var out bytes.Buffer

	err := handleSSE(reader, &out)
	if err != nil {
		t.Fatalf("handleSSE failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "This is plain text") {
		t.Errorf("expected output to contain plain text, got: %s", output)
	}
}

func TestIsEventStream(t *testing.T) {
	tests := []struct {
		contentType string
		expected    bool
	}{
		{"text/event-stream", true},
		{"text/event-stream; charset=utf-8", true},
		{"application/json", false},
		{"text/plain", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			result := isEventStream(tt.contentType)
			if result != tt.expected {
				t.Errorf("isEventStream(%q) = %v, want %v", tt.contentType, result, tt.expected)
			}
		})
	}
}

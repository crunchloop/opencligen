package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadBody_EmptyString(t *testing.T) {
	body, err := LoadBody("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != nil {
		t.Errorf("expected nil body, got %v", body)
	}
}

func TestLoadBody_RawJSON(t *testing.T) {
	input := `{"name": "test", "value": 123}`
	body, err := LoadBody(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != input {
		t.Errorf("expected %q, got %q", input, string(body))
	}
}

func TestLoadBody_FromFile(t *testing.T) {
	// Create a temp file with JSON content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	content := `{"key": "value"}`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	body, err := LoadBody("@" + testFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != content {
		t.Errorf("expected %q, got %q", content, string(body))
	}
}

func TestLoadBody_FileNotFound(t *testing.T) {
	_, err := LoadBody("@/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestLoadBody_RawText(t *testing.T) {
	// Non-JSON content should still work
	input := "plain text content"
	body, err := LoadBody(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != input {
		t.Errorf("expected %q, got %q", input, string(body))
	}
}

func TestLoadBody_ComplexJSON(t *testing.T) {
	input := `{
		"array": [1, 2, 3],
		"nested": {"a": "b"},
		"null": null,
		"bool": true
	}`
	body, err := LoadBody(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != input {
		t.Errorf("expected %q, got %q", input, string(body))
	}
}

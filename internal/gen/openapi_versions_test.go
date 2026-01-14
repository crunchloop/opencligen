package gen

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/crunchloop/opencligen/internal/plan"
	"github.com/crunchloop/opencligen/internal/spec"
)

// TestE2E_OpenAPI30_Bookmarks tests generation and building with an OpenAPI 3.0.3 spec
func TestE2E_OpenAPI30_Bookmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	// Load test spec
	ctx := context.Background()
	s, err := spec.Load(ctx, "../testdata/openapi30.yaml")
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}

	// Build plan
	p := plan.Build(s, "bookmarks", "github.com/example/bookmarks")

	// Create temp directory
	outDir := t.TempDir()

	// Generate
	gen := New(p, outDir)
	if err := gen.Generate(); err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	// Run go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = outDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, output)
	}

	// Build
	binaryPath := filepath.Join(outDir, "bookmarks")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/bookmarks")
	buildCmd.Dir = outDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	// Test root help
	t.Run("root help shows all groups", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("help command failed: %v", err)
		}

		helpText := string(output)

		expectedGroups := []string{"bookmarks", "folders", "export", "events"}
		for _, group := range expectedGroups {
			if !strings.Contains(helpText, group) {
				t.Errorf("expected help to contain '%s' group", group)
			}
		}
	})

	// Test bookmarks list has expected flags
	t.Run("bookmarks list has pagination flags", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "bookmarks", "list", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("bookmarks list help failed: %v", err)
		}

		helpText := string(output)

		expectedFlags := []string{"--tag", "--cursor", "--per-page"}
		for _, flag := range expectedFlags {
			if !strings.Contains(helpText, flag) {
				t.Errorf("expected bookmarks list help to contain '%s' flag", flag)
			}
		}
	})

	// Test bookmarks create requires idempotency key
	t.Run("bookmarks create has required header flag", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "bookmarks", "create", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("bookmarks create help failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "--idempotency-key") {
			t.Error("expected bookmarks create help to contain --idempotency-key flag")
		}
		if !strings.Contains(helpText, "--data") {
			t.Error("expected bookmarks create help to contain --data flag")
		}
	})

	// Test bookmarks get shows positional arg
	t.Run("bookmarks get shows positional", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "bookmarks", "get", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("bookmarks get help failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "<bookmarkId>") {
			t.Error("expected bookmarks get help to show <bookmarkId> positional")
		}
	})

	// Test bookmarks delete has boolean header flag
	t.Run("bookmarks delete has confirm flag", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "bookmarks", "delete", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("bookmarks delete help failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "--confirm-delete") {
			t.Error("expected bookmarks delete help to contain --confirm-delete flag")
		}
	})

	// Test export has format enum flag
	t.Run("export has format flag", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "export", "get", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("export get help failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "--format") {
			t.Error("expected export get help to contain --format flag")
		}
	})

	// Test events subscribe exists (SSE)
	t.Run("events subscribe exists", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "events", "subscribe", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("events subscribe help failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "Stream") {
			t.Error("expected events subscribe help to contain description")
		}
		if !strings.Contains(helpText, "--since") {
			t.Error("expected events subscribe help to contain --since flag")
		}
	})

	// Test folders list has boolean flag
	t.Run("folders list has include-count flag", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "folders", "list", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("folders list help failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "--include-count") {
			t.Error("expected folders list help to contain --include-count flag")
		}
	})
}

// TestE2E_OpenAPI31_Notes tests generation and building with an OpenAPI 3.1.0 spec
func TestE2E_OpenAPI31_Notes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	// Load test spec
	ctx := context.Background()
	s, err := spec.Load(ctx, "../testdata/openapi31.yaml")
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}

	// Build plan
	p := plan.Build(s, "notes", "github.com/example/notes")

	// Create temp directory
	outDir := t.TempDir()

	// Generate
	gen := New(p, outDir)
	if err := gen.Generate(); err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	// Run go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = outDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, output)
	}

	// Build
	binaryPath := filepath.Join(outDir, "notes")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/notes")
	buildCmd.Dir = outDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	// Test root help shows expected groups, not internal
	t.Run("root help shows groups but not internal", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("help command failed: %v", err)
		}

		helpText := string(output)

		expectedGroups := []string{"notes", "tags", "sync"}
		for _, group := range expectedGroups {
			if !strings.Contains(helpText, group) {
				t.Errorf("expected help to contain '%s' group", group)
			}
		}

		// internal should be hidden
		if strings.Contains(helpText, "internal") {
			t.Error("internal group should be hidden from help")
		}
	})

	// Test notes list has expected flags
	t.Run("notes list has filter flags", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "notes", "list", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("notes list help failed: %v", err)
		}

		helpText := string(output)

		expectedFlags := []string{"--status", "--search", "--sort", "--limit"}
		for _, flag := range expectedFlags {
			if !strings.Contains(helpText, flag) {
				t.Errorf("expected notes list help to contain '%s' flag", flag)
			}
		}
	})

	// Test notes get has include flag
	t.Run("notes get has include flag", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "notes", "get", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("notes get help failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "--include") {
			t.Error("expected notes get help to contain --include flag")
		}
		if !strings.Contains(helpText, "<noteId>") {
			t.Error("expected notes get help to show <noteId> positional")
		}
	})

	// Test notes update exists (PATCH)
	t.Run("notes update exists", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "notes", "update", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("notes update help failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "update") || !strings.Contains(helpText, "--data") {
			t.Error("expected notes update help to contain description and --data flag")
		}
	})

	// Test notes share-note has alias
	t.Run("notes share-note has alias", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "notes", "share-note", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("notes share-note help failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "Aliases") {
			t.Error("expected notes share-note help to show Aliases section")
		}
		if !strings.Contains(helpText, "sh") {
			t.Error("expected notes share-note help to show 'sh' alias")
		}
	})

	// Test tags list has shorthand
	t.Run("tags list has shorthand", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "tags", "list", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("tags list help failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "-m") {
			t.Error("expected tags list help to show -m shorthand")
		}
		if !strings.Contains(helpText, "--min-count") {
			t.Error("expected tags list help to contain --min-count flag")
		}
	})

	// Test sync subscribe has session flag with env hint
	t.Run("sync subscribe has session flag", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "sync", "subscribe", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("sync subscribe help failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "--session") {
			t.Error("expected sync subscribe help to contain --session flag")
		}
		if !strings.Contains(helpText, "-s") {
			t.Error("expected sync subscribe help to show -s shorthand")
		}
	})

	// Test notes delete has permanent flag
	t.Run("notes delete has permanent flag", func(t *testing.T) {
		output, err := exec.Command(binaryPath, "notes", "delete", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("notes delete help failed: %v", err)
		}

		helpText := string(output)

		if !strings.Contains(helpText, "--permanent") {
			t.Error("expected notes delete help to contain --permanent flag")
		}
	})

	// Test internal group still exists but is hidden
	t.Run("internal group exists but hidden", func(t *testing.T) {
		// The command should still work even though it's hidden
		// operationId getMetrics becomes "get" command
		output, err := exec.Command(binaryPath, "internal", "get", "--help").CombinedOutput()
		if err != nil {
			t.Fatalf("internal get help failed: %v", err)
		}

		helpText := string(output)

		// Just verify the command is accessible
		if !strings.Contains(helpText, "Internal") {
			t.Error("expected internal get to still be accessible")
		}
	})
}

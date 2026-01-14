package plan

import (
	"context"
	"testing"

	"github.com/crunchloop/opencligen/internal/spec"
)

func loadTestSpec(t *testing.T) *spec.Spec {
	t.Helper()
	ctx := context.Background()
	s, err := spec.Load(ctx, "../testdata/dap.json")
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}
	return s
}

func TestBuild_WorkspacesListHasPaginationFlags(t *testing.T) {
	s := loadTestSpec(t)
	plan := Build(s, "dap", "github.com/example/dap")

	// Find workspaces group
	var workspacesGroup *GroupPlan
	for i := range plan.Groups {
		if plan.Groups[i].Name == "workspaces" {
			workspacesGroup = &plan.Groups[i]
			break
		}
	}

	if workspacesGroup == nil {
		t.Fatal("expected to find workspaces group")
	}

	// Find list operation
	var listOp *OpPlan
	for i := range workspacesGroup.Operations {
		if workspacesGroup.Operations[i].CommandPath[1] == "list" {
			listOp = &workspacesGroup.Operations[i]
			break
		}
	}

	if listOp == nil {
		t.Fatal("expected to find list operation in workspaces")
	}

	// Check for --page and --limit flags
	var pageFlag, limitFlag *ParamPlan
	for i := range listOp.Flags {
		switch listOp.Flags[i].FlagName {
		case "page":
			pageFlag = &listOp.Flags[i]
		case "limit":
			limitFlag = &listOp.Flags[i]
		}
	}

	if pageFlag == nil {
		t.Error("expected --page flag")
	}

	if limitFlag == nil {
		t.Error("expected --limit flag")
	}
}

func TestBuild_WorkspacesGetHasPositionalId(t *testing.T) {
	s := loadTestSpec(t)
	plan := Build(s, "dap", "github.com/example/dap")

	// Find workspaces group
	var workspacesGroup *GroupPlan
	for i := range plan.Groups {
		if plan.Groups[i].Name == "workspaces" {
			workspacesGroup = &plan.Groups[i]
			break
		}
	}

	if workspacesGroup == nil {
		t.Fatal("expected to find workspaces group")
	}

	// Find get operation
	var getOp *OpPlan
	for i := range workspacesGroup.Operations {
		if workspacesGroup.Operations[i].CommandPath[1] == "get" {
			getOp = &workspacesGroup.Operations[i]
			break
		}
	}

	if getOp == nil {
		t.Fatal("expected to find get operation in workspaces")
	}

	// Check for positional id
	if len(getOp.Positionals) != 1 {
		t.Fatalf("expected 1 positional, got %d", len(getOp.Positionals))
	}

	if getOp.Positionals[0].Name != "id" {
		t.Errorf("expected positional named 'id', got '%s'", getOp.Positionals[0].Name)
	}
}

func TestBuild_TasksCreateHasUserIdAndDataFlags(t *testing.T) {
	s := loadTestSpec(t)
	plan := Build(s, "dap", "github.com/example/dap")

	// Find tasks group
	var tasksGroup *GroupPlan
	for i := range plan.Groups {
		if plan.Groups[i].Name == "tasks" {
			tasksGroup = &plan.Groups[i]
			break
		}
	}

	if tasksGroup == nil {
		t.Fatal("expected to find tasks group")
	}

	// Find create operation
	var createOp *OpPlan
	for i := range tasksGroup.Operations {
		if tasksGroup.Operations[i].CommandPath[1] == "create" {
			createOp = &tasksGroup.Operations[i]
			break
		}
	}

	if createOp == nil {
		t.Fatal("expected to find create operation in tasks")
	}

	// Check for --user-id flag (X-User-Id header)
	var userIDFlag *ParamPlan
	for i := range createOp.Flags {
		if createOp.Flags[i].FlagName == "user-id" {
			userIDFlag = &createOp.Flags[i]
			break
		}
	}

	if userIDFlag == nil {
		t.Error("expected --user-id flag (from X-User-Id header)")
	}

	// Check HasJSONBody is true (meaning --data will be generated)
	if !createOp.HasJSONBody {
		t.Error("expected HasJSONBody to be true")
	}
}

func TestBuild_StreamSubscribeIsDetectedAsStream(t *testing.T) {
	s := loadTestSpec(t)
	plan := Build(s, "dap", "github.com/example/dap")

	// Find stream group
	var streamGroup *GroupPlan
	for i := range plan.Groups {
		if plan.Groups[i].Name == "stream" {
			streamGroup = &plan.Groups[i]
			break
		}
	}

	if streamGroup == nil {
		t.Fatal("expected to find stream group")
	}

	// Find subscribe operation
	var subscribeOp *OpPlan
	for i := range streamGroup.Operations {
		if streamGroup.Operations[i].CommandPath[1] == "subscribe" {
			subscribeOp = &streamGroup.Operations[i]
			break
		}
	}

	if subscribeOp == nil {
		t.Fatal("expected to find subscribe operation in stream")
	}

	if !subscribeOp.IsEventStream {
		t.Error("expected stream subscribe to be detected as event stream")
	}
}

func TestDeriveCommandName(t *testing.T) {
	tests := []struct {
		operationID string
		expected    string
	}{
		{"listTasks", "list"},
		{"getTasks", "get"},
		{"createTask", "create"},
		{"updateTask", "update"},
		{"deleteTask", "delete"},
		{"startProcess", "start"},
		{"stopProcess", "stop"},
		{"cancelTask", "cancel"},
		{"pingHealth", "ping"},
		{"subscribeStream", "subscribe"},
		{"someCustomOperation", "some-custom-operation"},
		{"getUserById", "get"},
	}

	for _, tt := range tests {
		t.Run(tt.operationID, func(t *testing.T) {
			result := DeriveCommandName(tt.operationID)
			if result != tt.expected {
				t.Errorf("DeriveCommandName(%q) = %q, want %q", tt.operationID, result, tt.expected)
			}
		})
	}
}

func TestDeriveFlagName(t *testing.T) {
	tests := []struct {
		paramName string
		in        string
		expected  string
	}{
		{"page", "query", "page"},
		{"limit", "query", "limit"},
		{"X-User-Id", "header", "user-id"},
		{"X-Request-ID", "header", "request-id"},
		{"userId", "query", "user-id"},
		{"someParam", "query", "some-param"},
	}

	for _, tt := range tests {
		t.Run(tt.paramName, func(t *testing.T) {
			result := DeriveFlagName(tt.paramName, tt.in)
			if result != tt.expected {
				t.Errorf("DeriveFlagName(%q, %q) = %q, want %q", tt.paramName, tt.in, result, tt.expected)
			}
		})
	}
}

func TestParseCommandPath(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"tasks create", []string{"tasks", "create"}},
		{"tasks activities", []string{"tasks", "activities"}},
		{"users get profile", []string{"users", "get", "profile"}},
		{"TasksCreate", []string{"tasks-create"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseCommandPath(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("ParseCommandPath(%q) = %v, want %v", tt.input, result, tt.expected)
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("ParseCommandPath(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

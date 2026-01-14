package plan

import (
	"context"
	"testing"

	"github.com/crunchloop/opencligen/internal/spec"
)

func loadAnnotatedSpec(t *testing.T) *spec.Spec {
	t.Helper()
	ctx := context.Background()
	s, err := spec.Load(ctx, "../testdata/annotated.json")
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}
	return s
}

func TestXCli_CustomCommandName(t *testing.T) {
	s := loadAnnotatedSpec(t)
	plan := Build(s, "test", "github.com/example/test")

	// Find the task activities operation
	var activitiesOp *OpPlan
	for _, group := range plan.Groups {
		for i := range group.Operations {
			if group.Operations[i].OperationID == "listTaskActivities" {
				activitiesOp = &group.Operations[i]
				break
			}
		}
	}

	if activitiesOp == nil {
		t.Fatal("expected to find listTaskActivities operation")
	}

	// Check command path is ["tasks", "activities"]
	if len(activitiesOp.CommandPath) != 2 {
		t.Fatalf("expected 2 command path parts, got %d", len(activitiesOp.CommandPath))
	}

	if activitiesOp.CommandPath[0] != "tasks" {
		t.Errorf("expected first command path to be 'tasks', got '%s'", activitiesOp.CommandPath[0])
	}

	if activitiesOp.CommandPath[1] != "activities" {
		t.Errorf("expected second command path to be 'activities', got '%s'", activitiesOp.CommandPath[1])
	}
}

func TestXCli_Aliases(t *testing.T) {
	s := loadAnnotatedSpec(t)
	plan := Build(s, "test", "github.com/example/test")

	// Find the task activities operation
	var activitiesOp *OpPlan
	for _, group := range plan.Groups {
		for i := range group.Operations {
			if group.Operations[i].OperationID == "listTaskActivities" {
				activitiesOp = &group.Operations[i]
				break
			}
		}
	}

	if activitiesOp == nil {
		t.Fatal("expected to find listTaskActivities operation")
	}

	// Check aliases
	if len(activitiesOp.Aliases) != 2 {
		t.Fatalf("expected 2 aliases, got %d", len(activitiesOp.Aliases))
	}

	if activitiesOp.Aliases[0] != "act" {
		t.Errorf("expected first alias to be 'act', got '%s'", activitiesOp.Aliases[0])
	}

	if activitiesOp.Aliases[1] != "a" {
		t.Errorf("expected second alias to be 'a', got '%s'", activitiesOp.Aliases[1])
	}
}

func TestXCli_Hidden(t *testing.T) {
	s := loadAnnotatedSpec(t)
	plan := Build(s, "test", "github.com/example/test")

	// Find the internal sync operation
	var syncOp *OpPlan
	for _, group := range plan.Groups {
		for i := range group.Operations {
			if group.Operations[i].OperationID == "internalSync" {
				syncOp = &group.Operations[i]
				break
			}
		}
	}

	if syncOp == nil {
		t.Fatal("expected to find internalSync operation")
	}

	if !syncOp.Hidden {
		t.Error("expected operation to be hidden")
	}
}

func TestXCli_ParamFlagOverride(t *testing.T) {
	s := loadAnnotatedSpec(t)
	plan := Build(s, "test", "github.com/example/test")

	// Find the task activities operation
	var activitiesOp *OpPlan
	for _, group := range plan.Groups {
		for i := range group.Operations {
			if group.Operations[i].OperationID == "listTaskActivities" {
				activitiesOp = &group.Operations[i]
				break
			}
		}
	}

	if activitiesOp == nil {
		t.Fatal("expected to find listTaskActivities operation")
	}

	// Find the org flag (derived from X-Org-Id header)
	var orgFlag *ParamPlan
	for i := range activitiesOp.Flags {
		if activitiesOp.Flags[i].FlagName == "org" {
			orgFlag = &activitiesOp.Flags[i]
			break
		}
	}

	if orgFlag == nil {
		t.Fatal("expected to find org flag")
	}

	if orgFlag.Shorthand != "o" {
		t.Errorf("expected shorthand 'o', got '%s'", orgFlag.Shorthand)
	}

	if orgFlag.EnvVar != "ORG_ID" {
		t.Errorf("expected env var 'ORG_ID', got '%s'", orgFlag.EnvVar)
	}
}

func TestXCli_ParamShorthand(t *testing.T) {
	s := loadAnnotatedSpec(t)
	plan := Build(s, "test", "github.com/example/test")

	// Find the task activities operation
	var activitiesOp *OpPlan
	for _, group := range plan.Groups {
		for i := range group.Operations {
			if group.Operations[i].OperationID == "listTaskActivities" {
				activitiesOp = &group.Operations[i]
				break
			}
		}
	}

	if activitiesOp == nil {
		t.Fatal("expected to find listTaskActivities operation")
	}

	// Find the page flag
	var pageFlag *ParamPlan
	for i := range activitiesOp.Flags {
		if activitiesOp.Flags[i].FlagName == "page" {
			pageFlag = &activitiesOp.Flags[i]
			break
		}
	}

	if pageFlag == nil {
		t.Fatal("expected to find page flag")
	}

	if pageFlag.Shorthand != "p" {
		t.Errorf("expected shorthand 'p', got '%s'", pageFlag.Shorthand)
	}
}

func TestXCli_PositionalFalse(t *testing.T) {
	s := loadAnnotatedSpec(t)
	plan := Build(s, "test", "github.com/example/test")

	// Find the get user operation
	var getUserOp *OpPlan
	for _, group := range plan.Groups {
		for i := range group.Operations {
			if group.Operations[i].OperationID == "getUser" {
				getUserOp = &group.Operations[i]
				break
			}
		}
	}

	if getUserOp == nil {
		t.Fatal("expected to find getUser operation")
	}

	// userId should NOT be a positional (x-cli.positional = false)
	if len(getUserOp.Positionals) != 0 {
		t.Errorf("expected 0 positionals (userId marked as non-positional), got %d", len(getUserOp.Positionals))
	}

	// userId should be a flag instead
	var userFlag *ParamPlan
	for i := range getUserOp.Flags {
		if getUserOp.Flags[i].FlagName == "user" {
			userFlag = &getUserOp.Flags[i]
			break
		}
	}

	if userFlag == nil {
		t.Fatal("expected userId to be converted to --user flag")
	}
}

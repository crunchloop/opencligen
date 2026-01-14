package spec

import (
	"context"
	"testing"
)

func TestLoad(t *testing.T) {
	ctx := context.Background()
	spec, err := Load(ctx, "../testdata/dap.json")
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}

	if spec.Title != "DAP API" {
		t.Errorf("expected title 'DAP API', got '%s'", spec.Title)
	}

	if len(spec.Operations) == 0 {
		t.Fatal("expected operations to be extracted")
	}
}

func TestLoad_CreateTaskHasRequiredUserIdHeader(t *testing.T) {
	ctx := context.Background()
	spec, err := Load(ctx, "../testdata/dap.json")
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}

	var createTaskOp *Operation
	for i := range spec.Operations {
		if spec.Operations[i].OperationID == "createTask" {
			createTaskOp = &spec.Operations[i]
			break
		}
	}

	if createTaskOp == nil {
		t.Fatal("expected to find createTask operation")
	}

	// Check for X-User-Id header param
	var userIDParam *Param
	for i := range createTaskOp.Params {
		if createTaskOp.Params[i].Name == "X-User-Id" {
			userIDParam = &createTaskOp.Params[i]
			break
		}
	}

	if userIDParam == nil {
		t.Fatal("expected to find X-User-Id parameter")
	}

	if userIDParam.In != "header" {
		t.Errorf("expected X-User-Id to be in header, got '%s'", userIDParam.In)
	}

	if !userIDParam.Required {
		t.Error("expected X-User-Id to be required")
	}
}

func TestLoad_StreamEndpointHasEventStream(t *testing.T) {
	ctx := context.Background()
	spec, err := Load(ctx, "../testdata/dap.json")
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}

	var streamOp *Operation
	for i := range spec.Operations {
		if spec.Operations[i].OperationID == "subscribeStream" {
			streamOp = &spec.Operations[i]
			break
		}
	}

	if streamOp == nil {
		t.Fatal("expected to find subscribeStream operation")
	}

	if !streamOp.HasEventStream() {
		t.Error("expected stream endpoint to have text/event-stream content type")
	}

	// Verify the content type is present in responses
	found := false
	for _, resp := range streamOp.Responses {
		for _, ct := range resp.ContentTypes {
			if ct == "text/event-stream" {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("expected to find text/event-stream in response content types")
	}
}

func TestLoad_OperationsHaveCorrectTags(t *testing.T) {
	ctx := context.Background()
	spec, err := Load(ctx, "../testdata/dap.json")
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}

	tagCounts := make(map[string]int)
	for _, op := range spec.Operations {
		tagCounts[op.Tag]++
	}

	if tagCounts["tasks"] != 4 {
		t.Errorf("expected 4 tasks operations, got %d", tagCounts["tasks"])
	}

	if tagCounts["workspaces"] != 2 {
		t.Errorf("expected 2 workspaces operations, got %d", tagCounts["workspaces"])
	}

	if tagCounts["stream"] != 1 {
		t.Errorf("expected 1 stream operation, got %d", tagCounts["stream"])
	}
}

func TestLoad_PaginationParams(t *testing.T) {
	ctx := context.Background()
	spec, err := Load(ctx, "../testdata/dap.json")
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}

	var listWorkspacesOp *Operation
	for i := range spec.Operations {
		if spec.Operations[i].OperationID == "listWorkspaces" {
			listWorkspacesOp = &spec.Operations[i]
			break
		}
	}

	if listWorkspacesOp == nil {
		t.Fatal("expected to find listWorkspaces operation")
	}

	// Check for page and limit params
	var pageParam, limitParam *Param
	for i := range listWorkspacesOp.Params {
		switch listWorkspacesOp.Params[i].Name {
		case "page":
			pageParam = &listWorkspacesOp.Params[i]
		case "limit":
			limitParam = &listWorkspacesOp.Params[i]
		}
	}

	if pageParam == nil {
		t.Fatal("expected to find page parameter")
	}

	if limitParam == nil {
		t.Fatal("expected to find limit parameter")
	}

	if pageParam.In != "query" {
		t.Errorf("expected page to be query param, got '%s'", pageParam.In)
	}

	if limitParam.In != "query" {
		t.Errorf("expected limit to be query param, got '%s'", limitParam.In)
	}

	// Check min/max constraints on limit
	if limitParam.Min == nil || *limitParam.Min != 1 {
		t.Error("expected limit min to be 1")
	}

	if limitParam.Max == nil || *limitParam.Max != 100 {
		t.Error("expected limit max to be 100")
	}
}

func TestOperation_HasJSONBody(t *testing.T) {
	ctx := context.Background()
	spec, err := Load(ctx, "../testdata/dap.json")
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}

	var createTaskOp *Operation
	for i := range spec.Operations {
		if spec.Operations[i].OperationID == "createTask" {
			createTaskOp = &spec.Operations[i]
			break
		}
	}

	if createTaskOp == nil {
		t.Fatal("expected to find createTask operation")
	}

	if !createTaskOp.HasJSONBody() {
		t.Error("expected createTask to have JSON body")
	}
}

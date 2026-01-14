package spec

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Load loads and validates an OpenAPI spec from a file path
func Load(ctx context.Context, path string) (*Spec, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec: %w", err)
	}

	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("spec validation failed: %w", err)
	}

	return normalize(doc)
}

// normalize converts an OpenAPI document to our internal model
func normalize(doc *openapi3.T) (*Spec, error) {
	spec := &Spec{
		Title:       doc.Info.Title,
		Version:     doc.Info.Version,
		Description: doc.Info.Description,
	}

	// Parse global x-cli extensions
	if cli, ok := doc.Extensions["x-cli"]; ok {
		overrides, err := parseCliOverrides(cli)
		if err != nil {
			return nil, fmt.Errorf("failed to parse global x-cli: %w", err)
		}
		spec.GlobalCli = overrides
	}

	// Extract operations from paths
	// Sort paths for deterministic output
	paths := make([]string, 0, len(doc.Paths.Map()))
	for path := range doc.Paths.Map() {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		pathItem := doc.Paths.Map()[path]
		ops, err := extractOperations(path, pathItem)
		if err != nil {
			return nil, fmt.Errorf("failed to extract operations for path %s: %w", path, err)
		}
		spec.Operations = append(spec.Operations, ops...)
	}

	return spec, nil
}

// extractOperations extracts all operations from a path item
func extractOperations(path string, pathItem *openapi3.PathItem) ([]Operation, error) {
	var ops []Operation

	methods := []struct {
		method string
		op     *openapi3.Operation
	}{
		{"GET", pathItem.Get},
		{"POST", pathItem.Post},
		{"PUT", pathItem.Put},
		{"PATCH", pathItem.Patch},
		{"DELETE", pathItem.Delete},
		{"HEAD", pathItem.Head},
		{"OPTIONS", pathItem.Options},
	}

	for _, m := range methods {
		if m.op == nil {
			continue
		}

		op, err := extractOperation(path, m.method, m.op, pathItem.Parameters)
		if err != nil {
			return nil, fmt.Errorf("failed to extract %s operation: %w", m.method, err)
		}
		ops = append(ops, *op)
	}

	return ops, nil
}

// extractOperation extracts a single operation
func extractOperation(path, method string, op *openapi3.Operation, pathParams openapi3.Parameters) (*Operation, error) {
	tag := "default"
	if len(op.Tags) > 0 {
		tag = op.Tags[0]
	}

	operation := &Operation{
		Tag:         tag,
		Method:      method,
		Path:        path,
		OperationID: op.OperationID,
		Summary:     op.Summary,
		Description: op.Description,
	}

	// Parse operation-level x-cli
	if cli, ok := op.Extensions["x-cli"]; ok {
		overrides, err := parseCliOverrides(cli)
		if err != nil {
			return nil, fmt.Errorf("failed to parse operation x-cli: %w", err)
		}
		operation.Cli = overrides
	}

	// Extract parameters (path-level + operation-level)
	allParams := make([]*openapi3.ParameterRef, 0, len(pathParams)+len(op.Parameters))
	allParams = append(allParams, pathParams...)
	allParams = append(allParams, op.Parameters...)
	for _, paramRef := range allParams {
		if paramRef == nil || paramRef.Value == nil {
			continue
		}
		param, err := extractParam(paramRef.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to extract param %s: %w", paramRef.Value.Name, err)
		}
		operation.Params = append(operation.Params, *param)
	}

	// Extract request body
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		rb := op.RequestBody.Value
		reqBody := &RequestBody{
			Required:    rb.Required,
			Description: rb.Description,
		}
		for contentType := range rb.Content {
			reqBody.ContentTypes = append(reqBody.ContentTypes, contentType)
		}
		sort.Strings(reqBody.ContentTypes)
		operation.RequestBody = reqBody
	}

	// Extract responses
	if op.Responses != nil {
		// Sort status codes for deterministic output
		codes := make([]string, 0, len(op.Responses.Map()))
		for code := range op.Responses.Map() {
			codes = append(codes, code)
		}
		sort.Strings(codes)

		for _, code := range codes {
			respRef := op.Responses.Map()[code]
			if respRef == nil || respRef.Value == nil {
				continue
			}
			resp := respRef.Value

			response := Response{
				StatusCode:  code,
				Description: *resp.Description,
			}

			for contentType := range resp.Content {
				response.ContentTypes = append(response.ContentTypes, contentType)
			}
			sort.Strings(response.ContentTypes)

			operation.Responses = append(operation.Responses, response)
		}
	}

	return operation, nil
}

// extractParam extracts a parameter definition
func extractParam(p *openapi3.Parameter) (*Param, error) {
	param := &Param{
		Name:        p.Name,
		In:          p.In,
		Required:    p.Required,
		Description: p.Description,
	}

	// Extract type info from schema
	if p.Schema != nil && p.Schema.Value != nil {
		schema := p.Schema.Value
		param.Type = schema.Type.Slice()[0]
		param.Format = schema.Format
		param.Default = schema.Default

		if schema.Min != nil {
			param.Min = schema.Min
		}
		if schema.Max != nil {
			param.Max = schema.Max
		}
	}

	// Parse parameter-level x-cli
	if cli, ok := p.Extensions["x-cli"]; ok {
		overrides, err := parseParamCliOverrides(cli)
		if err != nil {
			return nil, fmt.Errorf("failed to parse param x-cli: %w", err)
		}
		param.Cli = overrides
	}

	return param, nil
}

// parseCliOverrides parses x-cli extensions at operation/global level
func parseCliOverrides(ext interface{}) (*CliOverrides, error) {
	data, err := json.Marshal(ext)
	if err != nil {
		return nil, err
	}
	var overrides CliOverrides
	if err := json.Unmarshal(data, &overrides); err != nil {
		return nil, err
	}
	return &overrides, nil
}

// parseParamCliOverrides parses x-cli extensions at parameter level
func parseParamCliOverrides(ext interface{}) (*ParamCliOverrides, error) {
	data, err := json.Marshal(ext)
	if err != nil {
		return nil, err
	}
	var overrides ParamCliOverrides
	if err := json.Unmarshal(data, &overrides); err != nil {
		return nil, err
	}
	return &overrides, nil
}

// HasEventStream checks if any response has text/event-stream content type
func (o *Operation) HasEventStream() bool {
	for _, resp := range o.Responses {
		for _, ct := range resp.ContentTypes {
			if strings.Contains(ct, "text/event-stream") {
				return true
			}
		}
	}
	return false
}

// HasJSONBody checks if the operation has a JSON request body
func (o *Operation) HasJSONBody() bool {
	if o.RequestBody == nil {
		return false
	}
	for _, ct := range o.RequestBody.ContentTypes {
		if strings.Contains(ct, "application/json") {
			return true
		}
	}
	return false
}

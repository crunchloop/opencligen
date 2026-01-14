// Package plan transforms OpenAPI specifications into CLI command plans.
//
// The plan package takes a parsed OpenAPI spec and builds a hierarchical
// command structure suitable for code generation. It handles:
//
//   - Grouping operations by OpenAPI tags
//   - Deriving command names from operationIds
//   - Converting parameters to CLI flags and positional arguments
//   - Applying x-cli overrides for customization
//   - Determining flag names, shorthands, and environment variables
//
// Example usage:
//
//	spec, _ := spec.Load(ctx, "api.json")
//	plan := plan.Build(spec, "mycli", "github.com/user/mycli")
//	fmt.Printf("Generated %d command groups\n", len(plan.Groups))
package plan

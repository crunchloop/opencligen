// Package spec provides OpenAPI 3.0 specification loading and parsing.
//
// The spec package reads OpenAPI specifications from JSON or YAML files,
// validates them, and normalizes them into internal data structures.
// It supports extraction of operations, parameters, request bodies,
// and response types, along with custom x-cli vendor extensions for
// CLI customization.
//
// Example usage:
//
//	ctx := context.Background()
//	spec, err := spec.Load(ctx, "api.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Loaded %d operations\n", len(spec.Operations))
package spec

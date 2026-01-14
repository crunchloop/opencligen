// Package gen generates complete Go CLI applications from command plans.
//
// The gen package takes a plan.Plan and produces a fully functional Go CLI
// application including:
//
//   - Main entry point (cmd/<app>/main.go)
//   - Root command with global flags
//   - Group commands for each tag
//   - Operation commands for each endpoint
//   - Runtime library for HTTP execution
//   - go.mod with required dependencies
//
// The generated code uses cobra for CLI structure and includes support for
// JSON request bodies, SSE streaming, and configuration file loading.
//
// Example usage:
//
//	plan := plan.Build(spec, "mycli", "github.com/user/mycli")
//	generator := gen.New(plan, "/path/to/output")
//	if err := generator.Generate(); err != nil {
//	    log.Fatal(err)
//	}
package gen

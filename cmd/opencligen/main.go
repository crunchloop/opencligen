package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/crunchloop/opencligen/internal/gen"
	"github.com/crunchloop/opencligen/internal/plan"
	"github.com/crunchloop/opencligen/internal/spec"
)

// Version information set by ldflags during build
var (
	version   = "dev"
	buildTime = "unknown"
)

var (
	specPath   string
	outDir     string
	appName    string
	moduleName string
	doBuild    bool
	dryRun     bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "opencligen",
		Short:   "Generate CLI tools from OpenAPI specifications",
		Version: version,
		Long: `opencligen generates a complete Go CLI application from an OpenAPI 3.0 spec.
The generated CLI includes one command per endpoint, grouped by tags,
with support for x-cli overrides for customizing names, flags, and configuration.`,
	}

	// Set version template to include build time
	rootCmd.SetVersionTemplate(fmt.Sprintf("opencligen version %s (built %s)\n", version, buildTime))

	genCmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate a CLI from an OpenAPI spec",
		Long: `Generate a Go CLI application from an OpenAPI 3.0 specification.

The generated CLI will have:
- One command per endpoint
- Commands grouped by tags
- Support for x-cli overrides
- JSON and SSE response handling`,
		RunE: runGen,
	}

	genCmd.Flags().StringVar(&specPath, "spec", "", "Path to OpenAPI spec file (required)")
	genCmd.Flags().StringVar(&outDir, "out", "", "Output directory (required)")
	genCmd.Flags().StringVar(&appName, "name", "", "Application name (required)")
	genCmd.Flags().StringVar(&moduleName, "module", "", "Go module name (optional, defaults to app name)")
	genCmd.Flags().BoolVar(&doBuild, "build", false, "Build the generated CLI after generation")
	genCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print plan without generating files")

	_ = genCmd.MarkFlagRequired("spec")
	_ = genCmd.MarkFlagRequired("out")
	_ = genCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(genCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runGen(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate spec path
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		return fmt.Errorf("spec file not found: %s", specPath)
	}

	// Load and validate spec
	fmt.Printf("Loading spec from %s...\n", specPath)
	s, err := spec.Load(ctx, specPath)
	if err != nil {
		return fmt.Errorf("failed to load spec: %w", err)
	}

	fmt.Printf("Loaded spec: %s v%s (%d operations)\n", s.Title, s.Version, len(s.Operations))

	// Set default module name
	if moduleName == "" {
		moduleName = appName
	}

	// Build plan
	fmt.Println("Building command plan...")
	p := plan.Build(s, appName, moduleName)

	if dryRun {
		printPlan(p)
		return nil
	}

	// Validate output directory
	outDir, err = filepath.Abs(outDir)
	if err != nil {
		return fmt.Errorf("invalid output directory: %w", err)
	}

	// Check if output directory is writable
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate
	fmt.Printf("Generating CLI to %s...\n", outDir)
	generator := gen.New(p, outDir)
	if err := generator.Generate(); err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	fmt.Println("Generation complete!")

	// Run go mod tidy
	fmt.Println("Running go mod tidy...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = outDir
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	// Build if requested
	if doBuild {
		fmt.Println("Building CLI...")
		binaryPath := filepath.Join(outDir, appName)
		buildCmd := exec.Command("go", "build", "-o", binaryPath, fmt.Sprintf("./cmd/%s", appName))
		buildCmd.Dir = outDir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}
		fmt.Printf("Built binary: %s\n", binaryPath)
	}

	fmt.Println("Done!")
	return nil
}

func printPlan(p *plan.Plan) {
	fmt.Printf("\n=== Command Plan for %s ===\n\n", p.AppName)
	fmt.Printf("Module: %s\n\n", p.ModuleName)

	for gi := range p.Groups {
		group := &p.Groups[gi]
		fmt.Printf("Group: %s\n", group.Name)
		for oi := range group.Operations {
			op := &group.Operations[oi]
			cmdPath := ""
			for i, part := range op.CommandPath {
				if i > 0 {
					cmdPath += " "
				}
				cmdPath += part
			}

			flags := ""
			for fi := range op.Flags {
				f := &op.Flags[fi]
				if flags != "" {
					flags += ", "
				}
				req := ""
				if f.Required {
					req = "*"
				}
				flags += fmt.Sprintf("--%s%s", f.FlagName, req)
			}

			positionals := ""
			for pi := range op.Positionals {
				pos := &op.Positionals[pi]
				positionals += fmt.Sprintf(" <%s>", pos.Name)
			}

			stream := ""
			if op.IsEventStream {
				stream = " [SSE]"
			}

			fmt.Printf("  %s%s%s\n", cmdPath, positionals, stream)
			if flags != "" {
				fmt.Printf("    Flags: %s\n", flags)
			}
			fmt.Printf("    %s %s\n", op.Method, op.Path)
		}
		fmt.Println()
	}
}

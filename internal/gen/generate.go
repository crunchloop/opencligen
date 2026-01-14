package gen

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/crunchloop/opencligen/internal/plan"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

//go:embed runtime/*.go
var runtimeFS embed.FS

// Generator generates a CLI from a plan
type Generator struct {
	Plan       *plan.Plan
	OutDir     string
	AppName    string
	ModuleName string
}

// New creates a new Generator
func New(p *plan.Plan, outDir string) *Generator {
	return &Generator{
		Plan:       p,
		OutDir:     outDir,
		AppName:    p.AppName,
		ModuleName: p.ModuleName,
	}
}

// Generate generates all files for the CLI
func (g *Generator) Generate() error {
	// Create output directories
	dirs := []string{
		filepath.Join(g.OutDir, "cmd", g.AppName),
		filepath.Join(g.OutDir, "internal", "runtime"),
		filepath.Join(g.OutDir, "internal", "commands"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate go.mod
	if err := g.generateGoMod(); err != nil {
		return fmt.Errorf("failed to generate go.mod: %w", err)
	}

	// Copy runtime files
	if err := g.copyRuntimeFiles(); err != nil {
		return fmt.Errorf("failed to copy runtime files: %w", err)
	}

	// Generate main.go
	if err := g.generateMain(); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}

	// Generate root.go
	if err := g.generateRoot(); err != nil {
		return fmt.Errorf("failed to generate root.go: %w", err)
	}

	// Generate group and operation files
	if err := g.generateCommands(); err != nil {
		return fmt.Errorf("failed to generate commands: %w", err)
	}

	return nil
}

func (g *Generator) generateGoMod() error {
	content := fmt.Sprintf(`module %s

go 1.22

require (
	github.com/spf13/cobra v1.8.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)
`, g.ModuleName)

	return os.WriteFile(filepath.Join(g.OutDir, "go.mod"), []byte(content), 0644)
}

func (g *Generator) copyRuntimeFiles() error {
	entries, err := runtimeFS.ReadDir("runtime")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		content, err := runtimeFS.ReadFile("runtime/" + entry.Name())
		if err != nil {
			return err
		}

		outPath := filepath.Join(g.OutDir, "internal", "runtime", entry.Name())
		if err := os.WriteFile(outPath, content, 0644); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateMain() error {
	tmpl, err := template.ParseFS(templateFS, "templates/cmd_main.go.tmpl")
	if err != nil {
		return err
	}

	data := map[string]string{
		"ModuleName": g.ModuleName,
		"AppName":    g.AppName,
	}

	return g.executeTemplate(tmpl, data, filepath.Join(g.OutDir, "cmd", g.AppName, "main.go"))
}

func (g *Generator) generateRoot() error {
	tmpl, err := template.ParseFS(templateFS, "templates/root.go.tmpl")
	if err != nil {
		return err
	}

	data := map[string]string{
		"ModuleName": g.ModuleName,
		"AppName":    g.AppName,
	}

	return g.executeTemplate(tmpl, data, filepath.Join(g.OutDir, "internal", "commands", "root.go"))
}

func (g *Generator) generateCommands() error {
	groupTmpl, err := template.ParseFS(templateFS, "templates/group.go.tmpl")
	if err != nil {
		return err
	}

	opTmpl, err := template.ParseFS(templateFS, "templates/operation.go.tmpl")
	if err != nil {
		return err
	}

	for _, group := range g.Plan.Groups {
		// Generate group file
		groupData := map[string]string{
			"VarName":     toVarName(group.Name),
			"Name":        group.Name,
			"Description": fmt.Sprintf("%s commands", capitalize(group.Name)),
		}

		groupFile := filepath.Join(g.OutDir, "internal", "commands", fmt.Sprintf("%s.go", group.Name))
		if err := g.executeTemplate(groupTmpl, groupData, groupFile); err != nil {
			return fmt.Errorf("failed to generate group %s: %w", group.Name, err)
		}

		// Generate operation files
		for _, op := range group.Operations {
			if err := g.generateOperation(opTmpl, group, op); err != nil {
				return fmt.Errorf("failed to generate operation %s: %w", op.OperationID, err)
			}
		}
	}

	return nil
}

func (g *Generator) generateOperation(tmpl *template.Template, group plan.GroupPlan, op plan.OpPlan) error {
	// Determine command name (last element of command path)
	cmdName := op.CommandPath[len(op.CommandPath)-1]

	// Build positionals data
	positionals := make([]map[string]interface{}, len(op.Positionals))
	for i, p := range op.Positionals {
		positionals[i] = map[string]interface{}{
			"Name":    p.Name,
			"VarName": toVarName(p.FlagName),
		}
	}

	// Build flags data
	flags := make([]map[string]interface{}, len(op.Flags))
	for i, p := range op.Flags {
		defaultStr := ""
		if p.Default != nil {
			defaultStr = fmt.Sprintf("%v", p.Default)
		}

		flags[i] = map[string]interface{}{
			"Name":        p.Name,
			"FlagName":    p.FlagName,
			"VarName":     toVarName(p.FlagName),
			"Type":        p.Type,
			"Required":    p.Required,
			"DefaultStr":  defaultStr,
			"Description": escapeDescription(p.Description),
			"Shorthand":   p.Shorthand,
			"EnvVar":      p.EnvVar,
			"In":          p.In,
		}
	}

	// Build use string with positionals
	use := cmdName
	for _, p := range op.Positionals {
		use += fmt.Sprintf(" <%s>", p.Name)
	}

	// Check if any flags are required
	hasRequiredFlags := false
	for _, p := range op.Flags {
		if p.Required {
			hasRequiredFlags = true
			break
		}
	}

	opVarName := toVarName(group.Name + "_" + cmdName)

	data := map[string]interface{}{
		"ModuleName":       g.ModuleName,
		"AppName":          g.AppName,
		"OpVarName":        opVarName,
		"ParentVarName":    toVarName(group.Name),
		"Use":              use,
		"Summary":          escapeDescription(op.Summary),
		"Description":      escapeDescription(op.Description),
		"Method":           op.Method,
		"Path":             op.Path,
		"Positionals":      positionals,
		"Flags":            flags,
		"HasJSONBody":      op.HasJSONBody,
		"IsEventStream":    op.IsEventStream,
		"Hidden":           op.Hidden,
		"Aliases":          op.Aliases,
		"HasRequiredFlags": hasRequiredFlags,
	}

	fileName := fmt.Sprintf("%s_%s.go", group.Name, cmdName)
	filePath := filepath.Join(g.OutDir, "internal", "commands", fileName)

	return g.executeTemplate(tmpl, data, filePath)
}

func (g *Generator) executeTemplate(tmpl *template.Template, data interface{}, outPath string) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	// Format the Go code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// If formatting fails, write unformatted for debugging
		if writeErr := os.WriteFile(outPath, buf.Bytes(), 0644); writeErr != nil {
			return writeErr
		}
		return fmt.Errorf("failed to format %s: %w", outPath, err)
	}

	return os.WriteFile(outPath, formatted, 0644)
}

// toVarName converts a kebab-case string to a valid Go variable name
func toVarName(s string) string {
	parts := strings.Split(s, "-")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = capitalize(parts[i])
		}
	}
	result := strings.Join(parts, "")

	// Handle underscore separators too
	parts = strings.Split(result, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = capitalize(parts[i])
		}
	}
	result = strings.Join(parts, "")

	// Ensure first character is lowercase for unexported variable
	if len(result) > 0 {
		runes := []rune(result)
		runes[0] = unicode.ToLower(runes[0])
		result = string(runes)
	}

	return result
}

// capitalize capitalizes the first letter of a string
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// escapeDescription escapes a string for use in Go code
func escapeDescription(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

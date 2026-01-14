package plan

import (
	"sort"

	"github.com/crunchloop/opencligen/internal/spec"
)

// Build creates a Plan from a Spec
func Build(s *spec.Spec, appName, moduleName string) *Plan {
	plan := &Plan{
		AppName:    appName,
		ModuleName: moduleName,
	}

	// Group operations by tag
	groups := make(map[string][]spec.Operation)
	for _, op := range s.Operations {
		tag := op.Tag
		if tag == "" {
			tag = "default"
		}
		groups[tag] = append(groups[tag], op)
	}

	// Sort group names for deterministic output
	groupNames := make([]string, 0, len(groups))
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	// Build group plans
	for _, groupName := range groupNames {
		ops := groups[groupName]
		groupPlan := buildGroupPlan(groupName, ops)
		plan.Groups = append(plan.Groups, groupPlan)
	}

	return plan
}

func buildGroupPlan(name string, ops []spec.Operation) GroupPlan {
	group := GroupPlan{
		Name: DeriveGroupName(name),
	}

	for _, op := range ops {
		opPlan := buildOpPlan(name, op)
		group.Operations = append(group.Operations, opPlan)
	}

	return group
}

func buildOpPlan(groupName string, op spec.Operation) OpPlan {
	opPlan := OpPlan{
		Method:        op.Method,
		Path:          op.Path,
		OperationID:   op.OperationID,
		Summary:       op.Summary,
		Description:   op.Description,
		HasJSONBody:   op.HasJSONBody(),
		IsEventStream: op.HasEventStream(),
	}

	// Determine command path
	if op.Cli != nil && op.Cli.Name != "" {
		opPlan.CommandPath = ParseCommandPath(op.Cli.Name)
	} else {
		// Default: [group, derived-command-name]
		cmdName := DeriveCommandName(op.OperationID)
		opPlan.CommandPath = []string{DeriveGroupName(groupName), cmdName}
	}

	// Apply operation-level x-cli overrides
	if op.Cli != nil {
		opPlan.Hidden = op.Cli.Hidden
		opPlan.Aliases = op.Cli.Aliases
		if op.Cli.Group != "" {
			// Override the group in the command path
			opPlan.CommandPath[0] = DeriveGroupName(op.Cli.Group)
		}
	}

	// Process parameters
	// First, collect path params to determine positional order
	var pathParams []spec.Param
	var otherParams []spec.Param

	for _, p := range op.Params {
		if p.In == "path" {
			pathParams = append(pathParams, p)
		} else {
			otherParams = append(otherParams, p)
		}
	}

	// Path params become positionals by default (in path order)
	for _, p := range pathParams {
		paramPlan := buildParamPlan(p)

		// Check if explicitly marked as non-positional
		isPositional := true
		if p.Cli != nil && p.Cli.Positional != nil {
			isPositional = *p.Cli.Positional
		}

		if isPositional {
			opPlan.Positionals = append(opPlan.Positionals, paramPlan)
		} else {
			opPlan.Flags = append(opPlan.Flags, paramPlan)
		}
	}

	// Other params become flags
	for _, p := range otherParams {
		paramPlan := buildParamPlan(p)
		opPlan.Flags = append(opPlan.Flags, paramPlan)
	}

	return opPlan
}

func buildParamPlan(p spec.Param) ParamPlan {
	plan := ParamPlan{
		Name:        p.Name,
		Type:        p.Type,
		Format:      p.Format,
		Required:    p.Required,
		Default:     p.Default,
		Min:         p.Min,
		Max:         p.Max,
		Description: p.Description,
		In:          p.In,
	}

	// Derive flag name
	if p.Cli != nil && p.Cli.Flag != "" {
		plan.FlagName = p.Cli.Flag
	} else {
		plan.FlagName = DeriveFlagName(p.Name, p.In)
	}

	// Apply other x-cli overrides
	if p.Cli != nil {
		plan.Shorthand = p.Cli.Shorthand
		plan.EnvVar = p.Cli.Env
		plan.ConfigKey = p.Cli.ConfigKey
	}

	return plan
}

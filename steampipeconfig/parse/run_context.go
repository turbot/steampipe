package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/stevenle/topsort"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type RunContext struct {
	Mod             *modconfig.Mod
	dependencyGraph *topsort.Graph
	dependencies    map[string]bool
	UnparsedBlocks  hcl.Blocks
}

// creation
func NewRunContext(mod *modconfig.Mod, blocks hcl.Blocks) *RunContext {
	return &RunContext{
		Mod:             mod,
		UnparsedBlocks:  blocks,
		dependencyGraph: topsort.NewGraph(),
		dependencies:    make(map[string]bool),
	}
}

// dependencies
func (c *RunContext) GetDependencyOrder(rootAction string) ([]string, error) {
	return c.dependencyGraph.TopSort(rootAction)
}

// mutation

func (c *RunContext) AddDependencies(name string, dependencies []hcl.Traversal) hcl.Diagnostics {

	// store dependency in tree - d
	if !c.dependencyGraph.ContainsNode(name) {
		c.dependencyGraph.AddNode(name)
	}
	for _, dep := range dependencies {
		d, diags := NameFromTraversal(dep)
		if diags.HasErrors() {
			return diags
		}
		if !c.dependencyGraph.ContainsNode(d) {
			c.dependencyGraph.AddNode(d)
		}

		// store the raw dependency properties
		c.dependencies[d] = true

		// HUYH>>>
		//c.dependencies.AddNode(d)
		c.dependencyGraph.AddEdge(name, d)

	}
	return nil
}

// state

// EvalComplete :: Are all elements in the dependency tree fully evaluated
func (c *RunContext) EvalComplete() bool {
	return len(c.UnparsedBlocks) == 0
}

// eval functions
func (c *RunContext) BuildEvalContext() (*hcl.EvalContext, error) {

	// build a variable map from the depends
	for dep := range c.dependencies {
		resourcePropertyPath, err := modconfig.ParseModResourcePropertyPath(dep)
		fmt.Println(resourcePropertyPath)
		if err != nil {
			return nil, err
		}
		// for now we only support locals
		if resourcePropertyPath.ItemType != modconfig.BlockTypeLocals {
			return nil, fmt.Errorf("could not resolve reference '%s' - in this version, only references to 'locals' are supported", dep)
		}
	}
	// create evaluation context
	ctx := &hcl.EvalContext{
		Variables: make(map[string]cty.Value),
		Functions: ContextFunctions(c.Mod.ModPath),
	}

	return ctx, nil
}

// StartEvalLoop :: start the evaluation loop
// clear UnparsedBlocks - they will get repopulated as we execute the loop
func (c *RunContext) StartEvalLoop() {
	c.UnparsedBlocks = nil
}

func (c *RunContext) AddUnparsedBlock(block *hcl.Block) {
	c.UnparsedBlocks = append(c.UnparsedBlocks, block)
}

package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type WorkspaceProfileParseContext struct {
	ParseContext
	workspaceProfiles map[string]*modconfig.WorkspaceProfile
	valueMap          map[string]cty.Value
}

func NewWorkspaceProfileParseContext(rootEvalPath string) *WorkspaceProfileParseContext {
	parseContext := NewParseContext(rootEvalPath)
	// TODO uncomment once https://github.com/turbot/steampipe/issues/2640 is done
	//parseContext.BlockTypes = []string{modconfig.BlockTypeWorkspaceProfile}
	c := &WorkspaceProfileParseContext{
		ParseContext:      parseContext,
		workspaceProfiles: make(map[string]*modconfig.WorkspaceProfile),
		valueMap:          make(map[string]cty.Value),
	}

	c.buildEvalContext()

	return c
}

// AddResource stores this resource as a variable to be added to the eval context. It alse
func (c *WorkspaceProfileParseContext) AddResource(workspaceProfile *modconfig.WorkspaceProfile) hcl.Diagnostics {
	ctyVal, err := workspaceProfile.CtyValue()
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to convert workspaceProfile '%s' to its cty value", workspaceProfile.ProfileName),
			Detail:   err.Error(),
			Subject:  &workspaceProfile.DeclRange,
		}}
	}

	c.workspaceProfiles[workspaceProfile.ProfileName] = workspaceProfile
	c.valueMap[workspaceProfile.ProfileName] = ctyVal

	// remove this resource from unparsed blocks
	delete(c.UnresolvedBlocks, workspaceProfile.ProfileName)

	c.buildEvalContext()

	return nil
}

func (c *WorkspaceProfileParseContext) buildEvalContext() {
	// rebuild the eval context
	// build a map with a single key - workspace
	vars := map[string]cty.Value{
		"workspace": cty.ObjectVal(c.valueMap),
	}
	c.ParseContext.buildEvalContext(vars)

}

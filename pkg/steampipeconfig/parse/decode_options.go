package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
)

// DecodeOptions decodes an options block
func DecodeOptions(block *hcl.Block, settings ...WithDecodeSetting) (options.Options, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	config := NewDecodeOptionsConfig()
	for _, applySetting := range settings {
		applySetting(config)
	}

	destination, ok := config.mapping[block.Labels[0]]
	if !ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Invalid options type '%s'", block.Labels[0]),
			Subject:  &block.DefRange,
		})
		return nil, diags
	}

	diags = gohcl.DecodeBody(block.Body, nil, destination)
	if diags.HasErrors() {
		return nil, diags
	}

	return destination, nil
}

type DecodeOptionsConfig struct {
	mapping map[string]options.Options
}
type WithDecodeSetting func(*DecodeOptionsConfig)

func NewDecodeOptionsConfig() *DecodeOptionsConfig {
	config := new(DecodeOptionsConfig)
	config.mapping = map[string]options.Options{
		options.ConnectionBlock: &options.Connection{},
		options.DatabaseBlock:   &options.Database{},
		options.TerminalBlock:   &options.Terminal{},
		options.GeneralBlock:    &options.General{},
		options.QueryBlock:      &options.Query{},
		options.CheckBlock:      &options.Check{},
		options.DashboardBlock:  &options.GlobalDashboard{},
	}
	return config
}

func AsWorkspaceProfileOption() WithDecodeSetting {
	return func(doc *DecodeOptionsConfig) {
		doc.mapping[options.DashboardBlock] = &options.WorkspaceProfileDashboard{}
	}
}

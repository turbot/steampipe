package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/hclhelpers"
	"github.com/turbot/pipe-fittings/v2/ociinstaller"
)

// MockPluginConnection is a mock implementation of PluginConnection for testing
type MockPluginConnection struct {
	name        string
	displayName string
	declRange   hclhelpers.Range
}

func (m *MockPluginConnection) GetDeclRange() hclhelpers.Range {
	return m.declRange
}

func (m *MockPluginConnection) GetName() string {
	return m.name
}

func (m *MockPluginConnection) GetDisplayName() string {
	return m.displayName
}

// TestPluginRemoveReport tests the PluginRemoveReport structure
func TestPluginRemoveReport(t *testing.T) {
	tests := map[string]struct {
		report   PluginRemoveReport
		validate func(*testing.T, PluginRemoveReport)
	}{
		"basic report": {
			report: PluginRemoveReport{
				Image:       ociinstaller.NewImageRef("hub.steampipe.io/plugins/turbot/aws@1.0.0"),
				ShortName:   "aws",
				Connections: []PluginConnection{},
			},
			validate: func(t *testing.T, report PluginRemoveReport) {
				assert.NotNil(t, report.Image)
				assert.Equal(t, "aws", report.ShortName)
				assert.Empty(t, report.Connections)
			},
		},
		"report with connections": {
			report: PluginRemoveReport{
				Image:     ociinstaller.NewImageRef("hub.steampipe.io/plugins/turbot/aws@1.0.0"),
				ShortName: "aws",
				Connections: []PluginConnection{
					&MockPluginConnection{
						name:        "aws",
						displayName: "aws",
						declRange: hclhelpers.Range{
							Filename: "config/aws.spc",
							Start:    hclhelpers.Pos{Line: 1, Column: 1},
							End:      hclhelpers.Pos{Line: 5, Column: 1},
						},
					},
					&MockPluginConnection{
						name:        "aws_prod",
						displayName: "aws_prod",
						declRange: hclhelpers.Range{
							Filename: "config/aws.spc",
							Start:    hclhelpers.Pos{Line: 7, Column: 1},
							End:      hclhelpers.Pos{Line: 12, Column: 1},
						},
					},
				},
			},
			validate: func(t *testing.T, report PluginRemoveReport) {
				assert.NotNil(t, report.Image)
				assert.Equal(t, "aws", report.ShortName)
				assert.Len(t, report.Connections, 2)
				assert.Equal(t, "aws", report.Connections[0].GetName())
				assert.Equal(t, "aws_prod", report.Connections[1].GetName())
			},
		},
		"report with connections in different files": {
			report: PluginRemoveReport{
				Image:     ociinstaller.NewImageRef("hub.steampipe.io/plugins/turbot/azure@2.0.0"),
				ShortName: "azure",
				Connections: []PluginConnection{
					&MockPluginConnection{
						name:        "azure",
						displayName: "azure",
						declRange: hclhelpers.Range{
							Filename: "config/azure.spc",
							Start:    hclhelpers.Pos{Line: 1, Column: 1},
							End:      hclhelpers.Pos{Line: 5, Column: 1},
						},
					},
					&MockPluginConnection{
						name:        "azure_prod",
						displayName: "azure_prod",
						declRange: hclhelpers.Range{
							Filename: "config/production.spc",
							Start:    hclhelpers.Pos{Line: 10, Column: 1},
							End:      hclhelpers.Pos{Line: 15, Column: 1},
						},
					},
				},
			},
			validate: func(t *testing.T, report PluginRemoveReport) {
				assert.NotNil(t, report.Image)
				assert.Equal(t, "azure", report.ShortName)
				assert.Len(t, report.Connections, 2)

				// Check different files
				file1 := report.Connections[0].GetDeclRange().Filename
				file2 := report.Connections[1].GetDeclRange().Filename
				assert.NotEqual(t, file1, file2)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.validate(t, tc.report)
		})
	}
}

// TestPluginRemoveReports tests the PluginRemoveReports type
func TestPluginRemoveReports(t *testing.T) {
	tests := map[string]struct {
		reports  PluginRemoveReports
		validate func(*testing.T, PluginRemoveReports)
	}{
		"empty reports": {
			reports: PluginRemoveReports{},
			validate: func(t *testing.T, reports PluginRemoveReports) {
				assert.Empty(t, reports)
			},
		},
		"single report": {
			reports: PluginRemoveReports{
				{
					Image:       ociinstaller.NewImageRef("hub.steampipe.io/plugins/turbot/aws@1.0.0"),
					ShortName:   "aws",
					Connections: []PluginConnection{},
				},
			},
			validate: func(t *testing.T, reports PluginRemoveReports) {
				assert.Len(t, reports, 1)
				assert.Equal(t, "aws", reports[0].ShortName)
			},
		},
		"multiple reports": {
			reports: PluginRemoveReports{
				{
					Image:       ociinstaller.NewImageRef("hub.steampipe.io/plugins/turbot/aws@1.0.0"),
					ShortName:   "aws",
					Connections: []PluginConnection{},
				},
				{
					Image:       ociinstaller.NewImageRef("hub.steampipe.io/plugins/turbot/azure@2.0.0"),
					ShortName:   "azure",
					Connections: []PluginConnection{},
				},
				{
					Image:       ociinstaller.NewImageRef("hub.steampipe.io/plugins/turbot/gcp@3.0.0"),
					ShortName:   "gcp",
					Connections: []PluginConnection{},
				},
			},
			validate: func(t *testing.T, reports PluginRemoveReports) {
				assert.Len(t, reports, 3)
				assert.Equal(t, "aws", reports[0].ShortName)
				assert.Equal(t, "azure", reports[1].ShortName)
				assert.Equal(t, "gcp", reports[2].ShortName)
			},
		},
		"reports with connections": {
			reports: PluginRemoveReports{
				{
					Image:     ociinstaller.NewImageRef("hub.steampipe.io/plugins/turbot/aws@1.0.0"),
					ShortName: "aws",
					Connections: []PluginConnection{
						&MockPluginConnection{
							name:        "aws",
							displayName: "aws",
							declRange: hclhelpers.Range{
								Filename: "config/aws.spc",
								Start:    hclhelpers.Pos{Line: 1, Column: 1},
								End:      hclhelpers.Pos{Line: 5, Column: 1},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, reports PluginRemoveReports) {
				assert.Len(t, reports, 1)
				assert.Len(t, reports[0].Connections, 1)
				assert.Equal(t, "aws", reports[0].Connections[0].GetName())
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.validate(t, tc.reports)
		})
	}
}

// TestPluginRemoveReportsPrint tests the Print method
func TestPluginRemoveReportsPrint(t *testing.T) {
	tests := map[string]struct {
		reports     PluginRemoveReports
		description string
	}{
		"print empty reports": {
			reports:     PluginRemoveReports{},
			description: "should not panic with empty reports",
		},
		"print single report without connections": {
			reports: PluginRemoveReports{
				{
					Image:       ociinstaller.NewImageRef("hub.steampipe.io/plugins/turbot/aws@1.0.0"),
					ShortName:   "aws",
					Connections: []PluginConnection{},
				},
			},
			description: "should print single plugin",
		},
		"print multiple reports": {
			reports: PluginRemoveReports{
				{
					Image:       ociinstaller.NewImageRef("hub.steampipe.io/plugins/turbot/aws@1.0.0"),
					ShortName:   "aws",
					Connections: []PluginConnection{},
				},
				{
					Image:       ociinstaller.NewImageRef("hub.steampipe.io/plugins/turbot/azure@2.0.0"),
					ShortName:   "azure",
					Connections: []PluginConnection{},
				},
			},
			description: "should print multiple plugins",
		},
		"print with connections": {
			reports: PluginRemoveReports{
				{
					Image:     ociinstaller.NewImageRef("hub.steampipe.io/plugins/turbot/aws@1.0.0"),
					ShortName: "aws",
					Connections: []PluginConnection{
						&MockPluginConnection{
							name:        "aws",
							displayName: "aws",
							declRange: hclhelpers.Range{
								Filename: "config/aws.spc",
								Start:    hclhelpers.Pos{Line: 1, Column: 1},
								End:      hclhelpers.Pos{Line: 5, Column: 1},
							},
						},
						&MockPluginConnection{
							name:        "aws_prod",
							displayName: "aws_prod",
							declRange: hclhelpers.Range{
								Filename: "config/aws.spc",
								Start:    hclhelpers.Pos{Line: 7, Column: 1},
								End:      hclhelpers.Pos{Line: 12, Column: 1},
							},
						},
					},
				},
			},
			description: "should print with connection details",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Test that Print doesn't panic
			assert.NotPanics(t, func() {
				tc.reports.Print()
			}, tc.description)
		})
	}
}

// TestMockPluginConnection removed - tests mock implementation instead of real code
// Mock tests provide no regression value for actual implementation
// Documented in cleanup report

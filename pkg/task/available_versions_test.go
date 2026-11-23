package task

import (
	"context"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/pipe-fittings/v2/versionfile"
)

func pluginReport(org, name, constraint, installed, latest string) plugin.PluginVersionCheckReport {
	r := plugin.PluginVersionCheckReport{Plugin: &versionfile.InstalledVersion{Version: installed}}
	r.CheckResponse.Org = org
	r.CheckResponse.Name = name
	r.CheckResponse.Constraint = constraint
	r.CheckResponse.Version = latest
	return r
}

// The update notice is drawn the same as by tablewriter v0.0.5: one bordered column, left-aligned, never wrapped, with
// its width taken from the text a person sees, so the colour codes in a line do not widen the box.
func TestAvailableVersionCacheAsTable(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = false
	t.Cleanup(func() { color.NoColor = noColor })
	mainVersion := viper.GetString("main.version")
	viper.Set("main.version", "2.5.2")
	t.Cleanup(func() { viper.Set("main.version", mainVersion) })

	tests := map[string]struct {
		cache *AvailableVersionCache
		want  string
	}{
		"cli": {
			cache: &AvailableVersionCache{CliCache: &CLIVersionCheckResponse{NewVersion: "2.6.0"}},
			want:  "+-------------------------------------------------------------------+\n|                                                                   |\n| A new version of Steampipe is available! \x1b[1m2.5.2\x1b[0m → \x1b[1m2.6.0\x1b[0m            |\n| You can update by downloading from \x1b[33mhttps://steampipe.io/downloads\x1b[0m |\n|                                                                   |\n+-------------------------------------------------------------------+\n",
		},
		"plugins": {
			cache: &AvailableVersionCache{PluginCache: map[string]plugin.PluginVersionCheckReport{
				"aws":   pluginReport("turbot", "aws", "latest", "1.2.0", "1.3.0"),
				"chaos": pluginReport("turbot", "chaos", "^0.4", "", "0.4.1"),
			}},
			want: "+-----------------------------------------------------------+\n|                                                           |\n| Updated versions of the following plugins are available:  |\n|                                                           |\n|   turbot/aws   @ latest           \x1b[1m     1.2.0\x1b[0m → \x1b[1m1.3.0     \x1b[0m |\n|   turbot/chaos @ ^0.4        →  \x1b[1m     0.4.1\x1b[0m                |\n|                                                           |\n| You can update by running \x1b[1msteampipe plugin update --all\x1b[0m   |\n|                                                           |\n+-----------------------------------------------------------+\n",
		},
		"cli and plugins": {
			cache: &AvailableVersionCache{
				CliCache: &CLIVersionCheckResponse{NewVersion: "2.6.0"},
				PluginCache: map[string]plugin.PluginVersionCheckReport{
					"net": pluginReport("turbot", "net", "latest", "1.0.0", "1.1.0"),
				},
			},
			want: "+-------------------------------------------------------------------+\n|                                                                   |\n| A new version of Steampipe is available! \x1b[1m2.5.2\x1b[0m → \x1b[1m2.6.0\x1b[0m            |\n| You can update by downloading from \x1b[33mhttps://steampipe.io/downloads\x1b[0m |\n|                                                                   |\n|                                                                   |\n| Updated versions of the following plugins are available:          |\n|                                                                   |\n|   turbot/net @ latest           \x1b[1m     1.0.0\x1b[0m → \x1b[1m1.1.0     \x1b[0m           |\n|                                                                   |\n| You can update by running \x1b[1msteampipe plugin update --all\x1b[0m           |\n|                                                                   |\n+-------------------------------------------------------------------+\n",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tc.cache.asTable(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if got == nil {
				t.Fatal("no table")
			}
			if got.String() != tc.want {
				t.Errorf("got:\n%s\nwant:\n%s", got.String(), tc.want)
			}
		})
	}
}

func TestAvailableVersionCacheAsTableEmpty(t *testing.T) {
	got, err := (&AvailableVersionCache{}).asTable(context.Background())
	if err != nil || got != nil {
		t.Fatalf("got %v, %v; want no table", got, err)
	}
}

package metaquery

import (
	"context"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/display"
)

func setOrGetSearchPath(ctx context.Context, input *HandlerInput) error {
	if len(input.args()) == 0 {
		sessionSearchPath := input.Client.GetRequiredSessionSearchPath()

		sessionSearchPath = helpers.RemoveFromStringSlice(sessionSearchPath, constants.InternalSchema)

		display.ShowWrappedTable(
			[]string{"search_path"},
			[][]string{
				{strings.Join(sessionSearchPath, ",")},
			},
			&display.ShowWrappedTableOptions{AutoMerge: false},
		)
	} else {
		arg := input.args()[0]
		var paths []string
		split := strings.Split(arg, ",")
		for _, s := range split {
			s = strings.TrimSpace(s)
			paths = append(paths, s)
		}
		viper.Set(constants.ArgSearchPath, paths)

		// now that the viper is set, call back into the client (exposed via QueryExecutor) which
		// already knows how to setup the search_paths with the viper values
		return input.Client.SetRequiredSessionSearchPath(ctx)
	}
	return nil
}

func setSearchPathPrefix(ctx context.Context, input *HandlerInput) error {
	arg := input.args()[0]
	paths := []string{}
	split := strings.Split(arg, ",")
	for _, s := range split {
		s = strings.TrimSpace(s)
		paths = append(paths, s)
	}
	viper.Set(constants.ArgSearchPathPrefix, paths)

	// now that the viper is set, call back into the client (exposed via QueryExecutor) which
	// already knows how to setup the search_paths with the viper values
	return input.Client.SetRequiredSessionSearchPath(ctx)
}

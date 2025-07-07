package metaquery

import (
	"context"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/querydisplay"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func setOrGetSearchPath(ctx context.Context, input *HandlerInput) error {
	if len(input.args()) == 0 {
		sessionSearchPath := input.Client.GetRequiredSessionSearchPath()

		sessionSearchPath = helpers.RemoveFromStringSlice(sessionSearchPath, constants.InternalSchema)

		querydisplay.ShowWrappedTable(
			[]string{"search_path"},
			[][]string{
				{strings.Join(sessionSearchPath, ",")},
			},
			&querydisplay.ShowWrappedTableOptions{AutoMerge: false},
		)
	} else {
		arg := input.args()[0]
		var paths []string
		split := strings.Split(arg, ",")
		for _, s := range split {
			s = strings.TrimSpace(s)
			paths = append(paths, s)
		}
		viper.Set(pconstants.ArgSearchPath, paths)

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
	viper.Set(pconstants.ArgSearchPathPrefix, paths)

	// now that the viper is set, call back into the client (exposed via QueryExecutor) which
	// already knows how to setup the search_paths with the viper values
	return input.Client.SetRequiredSessionSearchPath(ctx)
}

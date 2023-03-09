package controldisplay

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/karrick/gows"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
)

const MaxColumns = 200

type TextFormatter struct {
	FormatterBase
}

func (tf TextFormatter) Format(_ context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	renderer := NewTableRenderer(tree)
	widthConstraint := NewRangeConstraint(renderer.MinimumWidth(), MaxColumns)
	renderedText := renderer.Render(tf.getMaxCols(widthConstraint))
	res := strings.NewReader(fmt.Sprintf("\n%s\n", renderedText))
	return res, nil
}

func (tf TextFormatter) FileExtension() string {
	return constants.TextExtension
}

func (tf TextFormatter) Name() string {
	return constants.OutputFormatText
}

func (tf TextFormatter) Alias() string {
	return constants.OutputFormatBrief
}

func (tf TextFormatter) getMaxCols(constraint RangeConstraint) int {
	colsAvailable, _, _ := gows.GetWinSize()
	// check if STEAMPIPE_CHECK_DISPLAY_WIDTH env variable is set
	if viper.IsSet(constants.ArgCheckDisplayWidth) {
		colsAvailable = viper.GetInt(constants.ArgCheckDisplayWidth)
	}
	return constraint.Constrain(colsAvailable)
}

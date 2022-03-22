package controldisplay

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/karrick/gows"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/controlexecute"
)

const MaxColumns = 200

type TextFormatter struct{}

func (j *TextFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	renderer := NewTableRenderer(tree)
	widthConstraint := NewRangeConstraint(renderer.MinimumWidth(), MaxColumns)
	renderedText := renderer.Render(j.getMaxCols(widthConstraint))
	res := strings.NewReader(fmt.Sprintf("\n%s\n", renderedText))
	return res, nil
}

func (j *TextFormatter) FileExtension() string {
	return ".txt"
}

func (j *TextFormatter) getMaxCols(constraint RangeConstraint) int {
	colsAvailable, _, _ := gows.GetWinSize()
	// check if STEAMPIPE_CHECK_DISPLAY_WIDTH env variable is set
	if viper.IsSet(constants.ArgCheckDisplayWidth) {
		colsAvailable = viper.GetInt(constants.ArgCheckDisplayWidth)
	}
	return constraint.Constrain(colsAvailable)
}

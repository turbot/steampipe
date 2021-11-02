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
	return "txt"
}

func (j *TextFormatter) getMaxCols(constraint RangeConstraint) int {
	var colsAvailable int
	if viper.IsSet(constants.ArgCheckDisplayWidth) {
		colsAvailable = viper.GetInt(constants.ArgCheckDisplayWidth)
	} else {
		colsAvailable, _, _ = gows.GetWinSize()
	}
	return constraint.Constrain(colsAvailable)
}

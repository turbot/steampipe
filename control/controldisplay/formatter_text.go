package controldisplay

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/karrick/gows"
	"github.com/turbot/steampipe/control/execute"
)

const UsableMaxCols = 200

type TextFormatter struct{}

func (j *TextFormatter) Format(ctx context.Context, tree *execute.ExecutionTree) (io.Reader, error) {
	// limit to 200
	maxCols := j.getMaxCols(UsableMaxCols)
	renderer := NewTableRenderer(tree, maxCols)
	return (strings.NewReader(fmt.Sprintf("\n%s\n", renderer.Render()))), nil
}

func (j *TextFormatter) getMaxCols(limitCol int) int {
	maxCols, _, _ := gows.GetWinSize()
	if maxCols > limitCol {
		maxCols = limitCol
	}
	return maxCols
}

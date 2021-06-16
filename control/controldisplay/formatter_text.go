package controldisplay

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/karrick/gows"
	"github.com/turbot/steampipe/control/execute"
)

// limit text width
const maxCols = 200

type TextFormatter struct{}

func (j *TextFormatter) Format(ctx context.Context, tree *execute.ExecutionTree) (io.Reader, error) {
	maxCols := j.getMaxCols(maxCols)
	renderedText := NewTableRenderer(tree, maxCols).Render()
	res := strings.NewReader(fmt.Sprintf("\n%s\n", renderedText))
	return res, nil
}

func (j *TextFormatter) getMaxCols(limitCol int) int {
	maxCols, _, _ := gows.GetWinSize()
	if maxCols > limitCol {
		maxCols = limitCol
	}
	return maxCols
}

package controldisplay

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/karrick/gows"
	"github.com/turbot/steampipe/control/controlexecute"
)

type TextFormatter struct{}

func (j *TextFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	renderer := NewTableRenderer(tree)
	renderedText := renderer.Render(j.getMaxCols(renderer.MaxDepth()))
	res := strings.NewReader(fmt.Sprintf("\n%s\n", renderedText))
	return res, nil
}

func (j *TextFormatter) FileExtension() string {
	return "txt"
}

func (j *TextFormatter) getMaxCols(depth int) int {
	minimumWidthRequired := (depth * 2) + 8 /* severity */ + 6 /* counter */ + 12 /* counter graph */
	colRange := NewRange(minimumWidthRequired, 200)
	maxCols, _, _ := gows.GetWinSize()
	return colRange.Constrain(maxCols)
}

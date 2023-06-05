package controldisplay

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/utils"
)

const MaxColumns = 200

type TextFormatter struct {
	FormatterBase
}

func (tf TextFormatter) Format(_ context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	renderer := NewTableRenderer(tree)
	widthConstraint := utils.NewRangeConstraint(renderer.MinimumWidth(), MaxColumns)
	renderedText := renderer.Render(display.GetMaxCols(widthConstraint))
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

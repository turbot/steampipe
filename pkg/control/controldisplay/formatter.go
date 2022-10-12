package controldisplay

import (
	"context"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
	"io"
)

type Formatter interface {
	Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error)
	FileExtension() string
	Name() string
	Alias() string
	IsDefaultFormatterForExtension() bool
}

type FormatterBase struct{}

func (*FormatterBase) Alias() string {
	return ""
}
func (*FormatterBase) IsDefaultFormatterForExtension() bool {
	return false
}

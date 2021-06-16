package controldisplay

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/turbot/steampipe/control/execute"
)

type JSONFormatter struct{}

func (j *JSONFormatter) Format(ctx context.Context, tree *execute.ExecutionTree) (io.Reader, error) {
	bytes, err := json.MarshalIndent(tree.Root, "", "  ")
	if err != nil {
		return nil, err
	}
	res := strings.NewReader(string(bytes))
	return res, nil
}

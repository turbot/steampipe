package controldisplay

import (
	"context"
	"encoding/json"
	"io"

	"github.com/turbot/steampipe/control/controlexecute"
)

type JSONFormatter struct{}

func (j *JSONFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	reader, writer := io.Pipe()
	encoder := json.NewEncoder(writer)
	encoder.SetIndent(" ", " ")
	go func() {
		err := encoder.Encode(tree.Root)
		if err != nil {
			writer.CloseWithError(err)
			return
		}
		writer.Close()
	}()
	return reader, nil
}

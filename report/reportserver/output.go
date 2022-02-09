package reportserver

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/statushooks"
)

func outputEvent(ctx context.Context, msg string) {
	statushooks.Message(ctx, fmt.Sprintf("%s %s", "[ Event ]", msg))
}
func outputError(ctx context.Context, err error) {
	statushooks.Message(ctx, fmt.Sprintf("%s %s", "[ Error ]", err.Error()))
}

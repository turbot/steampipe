package utils

import (
	"context"
	"fmt"
	"strings"
)

// UserConfirmation displays the warning message and asks the user for input
// regarding whether to continue or not
func UserConfirmation(ctx context.Context, warningMsg string) (bool, error) {
	fmt.Println(warningMsg)
	confirm := make(chan string, 1)
	confirmErr := make(chan error, 1)

	go func() {
		defer func() {
			close(confirm)
			close(confirmErr)
		}()
		var userConfirm string
		_, err := fmt.Scanf("%s", &userConfirm)
		if err != nil {
			confirmErr <- err
			return
		}
		confirm <- userConfirm
	}()
	select {
	case err := <-confirmErr:
		return false, err
	case <-ctx.Done():
		return false, ctx.Err()
	case c := <-confirm:
		return strings.ToUpper(c) == "Y", nil
	}
}

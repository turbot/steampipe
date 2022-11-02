package error_helpers

import "errors"

var MissingCloudTokenError = errors.New("Not authenticated for Steampipe Cloud.\nPlease run 'steampipe login' or setup a token.")

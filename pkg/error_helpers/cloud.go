package error_helpers

func IsBadWorkspaceDatabaseArg(err error) bool {
	return err != nil && err.Error() == "404 Not Found"
}

func IsInvalidCloudToken(err error) bool {
	return err != nil && err.Error() == "401 Unauthorized"
}

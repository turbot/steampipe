package utils

import (
	"fmt"
	"strconv"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
)

func Base36Hash(input string, length int) (string, error) {
	// Get hex of the hash
	bs := helpers.GetMD5Hash(input)

	// Convert the first 16 chars of the hash from hex to base 36
	u1Hex := bs[0:16]
	u1, err := strconv.ParseUint(u1Hex, 16, 64)
	if err != nil {
		return "", sperr.WrapWithMessage(err, "unable to create hash.")
	}
	u1Base36 := strconv.FormatUint(u1, 36)

	// Either take the last {length} chars, or pad the result if needed
	if len(u1Base36) > length {
		return u1Base36[len(u1Base36)-length:], nil
	} else {
		return fmt.Sprintf("%0*s", length, u1Base36), nil
	}
}

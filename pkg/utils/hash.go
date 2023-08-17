package utils

import (
	"crypto/md5"
	"fmt"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
)

func Base36Hash(input string, length int) (string, error) {
	// Get a hash of it.
	// TODO - should this be sha1 or something else instead?
	h := md5.New()
	_, err := h.Write([]byte(input))
	if err != nil {
		return "", sperr.New("Unable to create hash.")
	}
	// Get hex of the hash
	bs := fmt.Sprintf("%x", h.Sum(nil))

	// Convert the first 16 chars of the hash from hex to base 36
	u1Hex := bs[0:16]
	u1, err := strconv.ParseUint(u1Hex, 16, 64)
	if err != nil {
		return "", sperr.New("Unable to create hash.")
	}
	u1Base36 := strconv.FormatUint(u1, 36)

	// Either take the last {length} chars, or pad the result if needed
	if len(u1Base36) > length {
		return u1Base36[len(u1Base36)-length:], nil
	} else {
		return fmt.Sprintf("%0*s", length, u1Base36), nil
	}
}

package utils

import (
	"fmt"
	"log"
	"strings"
)

// UserConfirmation asks the user for input regarding whether to continue or not
func UserConfirmation() bool {
	var userConfirm string
	_, err := fmt.Scanf("%s", &userConfirm)
	if err != nil {
		log.Fatal(err)
	}
	return strings.ToUpper(userConfirm) == "Y"
}

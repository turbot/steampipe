package utils

import (
	"fmt"
	"log"
	"strings"
)

// UserConfirmation displays the warning message and asks the user for input
// regarding whether to continue or not
func UserConfirmation(warning string) bool {
	fmt.Println(warning)
	var userConfirm string
	_, err := fmt.Scanf("%s", &userConfirm)
	if err != nil {
		log.Fatal(err)
	}
	return strings.ToUpper(userConfirm) == "Y"
}

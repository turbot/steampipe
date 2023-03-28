package utils

import (
	"fmt"
	"log"
)

// UserConfirmation asks the user for input regarding whether to continue or not
func UserConfirmation() bool {
	var userConfirm rune
	_, err := fmt.Scanf("%c", &userConfirm)
	if err != nil {
		log.Fatal(err)
	}
	keepMod := userConfirm == 'Y' || userConfirm == 'y'
	return keepMod
}

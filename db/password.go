package db

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
)

// Passwords :: structure for working with DB passwords
type Passwords struct {
	Root      string
	Steampipe string
}

func writePasswordFile(steampipePassword string, rootPassword string) error {
	log.Println("[TRACE]", "Writing password file")
	passwords := Passwords{
		Root:      rootPassword,
		Steampipe: steampipePassword,
	}
	content, err := json.Marshal(passwords)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(getPasswordFileLocation(), content, 0600)
}

func getPasswords() (*Passwords, error) {
	contentBytes, err := ioutil.ReadFile(getPasswordFileLocation())
	if err != nil {
		return nil, err
	}
	var passwords = new(Passwords)
	err = json.Unmarshal(contentBytes, passwords)
	if err != nil {
		return nil, err
	}
	return passwords, nil
}

func generatePassword(passwordLength, minSpecialChar, minNum, minUpperCase int) string {
	const (
		lowerCharSet   = "abcdedfghijklmnopqrst"
		upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		specialCharSet = "#$*"
		numberSet      = "0123456789"
		allCharSet     = lowerCharSet + upperCharSet + specialCharSet + numberSet
	)

	var password strings.Builder

	// set special character
	for i := 0; i < minSpecialChar; i++ {
		random := rand.Intn(len(specialCharSet))
		password.WriteString(string(specialCharSet[random]))
	}

	// set numeric
	for i := 0; i < minNum; i++ {
		random := rand.Intn(len(numberSet))
		password.WriteString(string(numberSet[random]))
	}

	// set uppercase
	for i := 0; i < minUpperCase; i++ {
		random := rand.Intn(len(upperCharSet))
		password.WriteString(string(upperCharSet[random]))
	}

	remainingLength := passwordLength - minSpecialChar - minNum - minUpperCase
	for i := 0; i < remainingLength; i++ {
		random := rand.Intn(len(allCharSet))
		password.WriteString(string(allCharSet[random]))
	}
	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})
	return string(inRune)
}

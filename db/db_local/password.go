package db_local

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/google/uuid"
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

func generatePassword() string {
	// Create a simple, random password of the form f9fe-442f-90fb
	// Simple to read / write, and has a strength rating of 4 per https://lowe.github.io/tryzxcvbn/
	// Yes, this UUIDv4 does always include a 4, but good enough for our needs.
	u, err := uuid.NewRandom()
	if err != nil {
		// Should never happen?
		panic(err)
	}
	s := u.String()
	return s[9:23]
}

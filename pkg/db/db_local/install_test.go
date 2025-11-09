package db_local

import "testing"

func TestIsValidDatabaseName(t *testing.T) {
	tests := map[string]bool{
		"valid_name":  true,
		"_valid_name": true,
		"InvalidName": false,
		"123Invalid":  false,
		"":            false, // empty string should return false
	}

	for dbName, expectedResult := range tests {
		if actualResult := isValidDatabaseName(dbName); actualResult != expectedResult {
			t.Logf("Expected %t for %s, but for %t", expectedResult, dbName, actualResult)
			t.Fail()
		}
	}
}

package db_local

import (
	"testing"
)

func TestIsValidDatabaseName(t *testing.T) {
	tests := map[string]bool{
		"valid_name":  true,
		"_valid_name": true,
		"InvalidName": false,
		"123Invalid":  false,
	}

	for dbName, expectedResult := range tests {
		if actualResult := isValidDatabaseName(dbName); actualResult != expectedResult {
			t.Logf("Expected %t for %s, but for %t", expectedResult, dbName, actualResult)
			t.Fail()
		}
	}
}

func TestIsValidDatabaseName_EmptyString(t *testing.T) {
	// Test that isValidDatabaseName handles empty strings gracefully
	// An empty string should return false, not panic
	result := isValidDatabaseName("")
	if result != false {
		t.Errorf("Expected false for empty string, got %v", result)
	}
}


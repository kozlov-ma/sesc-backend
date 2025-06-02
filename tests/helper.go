package tests

import (
	"fmt"
	"os"
	"testing"
)

// SkipIfNoTestAPI skips the integration test if TEST_API_URL is not set
func SkipIfNoTestAPI(t *testing.T) {
	t.Helper()

	if os.Getenv("TEST_API_URL") == "" {
		t.Skip("Skipping integration test, set TEST_API_URL to the test API server address")
	}
}

// CreateValidUserData creates a CreateUserRequest with all required fields filled with valid data
func CreateValidUserData(firstName, lastName string, role int) CreateUserRequest {
	return CreateUserRequest{
		FirstName:         firstName,
		LastName:          lastName,
		MiddleName:        "",
		Role:              role,
		PictureURL:        "",
		Subdivision:       "Test Subdivision",
		JobTitle:          "Test Position",
		EmploymentRate:    1.0,
		AcademicDegree:    0,
		PersonnelCategory: 1,
		EmploymentType:    1,
		AcademicTitle:     "",
		Honors:            "",
		Category:          "",
	}
}

// CreateValidUserDataWithSuffix creates a CreateUserRequest with unique names and valid data
func CreateValidUserDataWithSuffix(suffix string, role int) CreateUserRequest {
	return CreateValidUserData(
		fmt.Sprintf("Test_%s", suffix),
		fmt.Sprintf("User_%s", suffix),
		role,
	)
}

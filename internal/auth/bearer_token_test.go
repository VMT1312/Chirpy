package auth

import (
	"net/http"
	"testing"
)

func TestGetBearerToken(t *testing.T) {
	t.Run("valid bearer token", func(t *testing.T) {
		headers := http.Header{}
		expectedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
		headers.Set("Authorization", "Bearer "+expectedToken)

		token, err := GetBearerToken(headers)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if token != expectedToken {
			t.Errorf("Expected token %s, got %s", expectedToken, token)
		}
	})

	t.Run("no authorization header", func(t *testing.T) {
		headers := http.Header{}

		_, err := GetBearerToken(headers)

		if err == nil {
			t.Fatal("Expected error for missing authorization header")
		}

		expectedError := "no authorization header provided"
		if err.Error() != expectedError {
			t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
		}
	})
}

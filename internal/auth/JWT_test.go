package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	// Test data
	userID := uuid.New()
	tokenSecret := "test-secret-key"
	expiresIn := time.Hour

	t.Run("successful JWT creation", func(t *testing.T) {
		token, err := MakeJWT(userID, tokenSecret, expiresIn)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if token == "" {
			t.Fatal("Expected non-empty token")
		}

		// Verify the token can be parsed
		parsedToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		})

		if err != nil {
			t.Fatalf("Failed to parse generated token: %v", err)
		}

		if !parsedToken.Valid {
			t.Fatal("Generated token is not valid")
		}

		// Verify claims
		claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
		if !ok {
			t.Fatal("Failed to extract claims from token")
		}

		if claims.Issuer != "chirpy" {
			t.Errorf("Expected issuer 'chirpy', got %s", claims.Issuer)
		}

		if claims.Subject != userID.String() {
			t.Errorf("Expected subject %s, got %s", userID.String(), claims.Subject)
		}

		// Check expiration time is approximately correct (within 1 second tolerance)
		expectedExpiry := time.Now().Add(expiresIn)
		actualExpiry := claims.ExpiresAt.Time
		timeDiff := actualExpiry.Sub(expectedExpiry)
		if timeDiff > time.Second || timeDiff < -time.Second {
			t.Errorf("Expected expiry around %v, got %v", expectedExpiry, actualExpiry)
		}
	})

	t.Run("different user IDs produce different tokens", func(t *testing.T) {
		userID1 := uuid.New()
		userID2 := uuid.New()

		token1, err1 := MakeJWT(userID1, tokenSecret, expiresIn)
		token2, err2 := MakeJWT(userID2, tokenSecret, expiresIn)

		if err1 != nil || err2 != nil {
			t.Fatalf("Expected no errors, got %v, %v", err1, err2)
		}

		if token1 == token2 {
			t.Error("Expected different tokens for different user IDs")
		}
	})

	t.Run("different secrets produce different tokens", func(t *testing.T) {
		secret1 := "secret1"
		secret2 := "secret2"

		token1, err1 := MakeJWT(userID, secret1, expiresIn)
		token2, err2 := MakeJWT(userID, secret2, expiresIn)

		if err1 != nil || err2 != nil {
			t.Fatalf("Expected no errors, got %v, %v", err1, err2)
		}

		if token1 == token2 {
			t.Error("Expected different tokens for different secrets")
		}
	})

	t.Run("zero expiration time", func(t *testing.T) {
		token, err := MakeJWT(userID, tokenSecret, 0)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Token should be created but immediately expired
		parsedToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		})

		// The token should parse but might not be valid due to expiration
		if err != nil && err.Error() != "token is expired" && err.Error() != "token has invalid claims: token is expired" {
			t.Fatalf("Expected token to be expired or valid, got error: %v", err)
		}

		if parsedToken != nil {
			claims := parsedToken.Claims.(*jwt.RegisteredClaims)
			if !claims.ExpiresAt.Time.Before(time.Now()) && !claims.ExpiresAt.Time.Equal(time.Now()) {
				t.Error("Expected token to be expired or expire immediately")
			}
		}
	})
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret-key"
	expiresIn := time.Hour

	t.Run("validate valid JWT", func(t *testing.T) {
		// First create a valid token
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Now validate it
		validatedUserID, err := ValidateJWT(token, tokenSecret)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if validatedUserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, validatedUserID)
		}
	})

	t.Run("invalid token string", func(t *testing.T) {
		invalidToken := "invalid.token.string"

		_, err := ValidateJWT(invalidToken, tokenSecret)

		if err == nil {
			t.Fatal("Expected error for invalid token")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		// Create token with one secret
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Try to validate with different secret
		wrongSecret := "wrong-secret"
		_, err = ValidateJWT(token, wrongSecret)

		if err == nil {
			t.Fatal("Expected error when using wrong secret")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		// Create token that expires immediately
		expiredToken, err := MakeJWT(userID, tokenSecret, -time.Hour)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Try to validate expired token
		_, err = ValidateJWT(expiredToken, tokenSecret)

		if err == nil {
			t.Fatal("Expected error for expired token")
		}
	})

	t.Run("malformed token", func(t *testing.T) {
		malformedTokens := []string{
			"",
			"not.a.jwt",
			"header.payload", // Missing signature
			"too.many.parts.here.error",
		}

		for _, malformedToken := range malformedTokens {
			_, err := ValidateJWT(malformedToken, tokenSecret)
			if err == nil {
				t.Errorf("Expected error for malformed token: %s", malformedToken)
			}
		}
	})

	t.Run("token with invalid UUID in subject", func(t *testing.T) {
		// Create a custom token with invalid UUID in subject
		claims := jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Subject:   "invalid-uuid-string",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(tokenSecret))
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		_, err = ValidateJWT(tokenString, tokenSecret)

		if err == nil {
			t.Fatal("Expected error for token with invalid UUID in subject")
		}
	})

	t.Run("empty token string", func(t *testing.T) {
		_, err := ValidateJWT("", tokenSecret)

		if err == nil {
			t.Fatal("Expected error for empty token string")
		}
	})

	t.Run("empty secret", func(t *testing.T) {
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		_, err = ValidateJWT(token, "")

		if err == nil {
			t.Fatal("Expected error when validating with empty secret")
		}
	})
}

func TestMakeJWTAndValidateJWTIntegration(t *testing.T) {
	t.Run("round trip test", func(t *testing.T) {
		userID := uuid.New()
		tokenSecret := "integration-test-secret"
		expiresIn := time.Minute * 30

		// Create token
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Validate token
		validatedUserID, err := ValidateJWT(token, tokenSecret)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		// Verify the user ID matches
		if validatedUserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, validatedUserID)
		}
	})

	t.Run("multiple users round trip", func(t *testing.T) {
		tokenSecret := "multi-user-test-secret"
		expiresIn := time.Hour

		// Create multiple users and tokens
		users := []uuid.UUID{
			uuid.New(),
			uuid.New(),
			uuid.New(),
		}

		tokens := make([]string, len(users))

		// Create tokens for all users
		for i, userID := range users {
			token, err := MakeJWT(userID, tokenSecret, expiresIn)
			if err != nil {
				t.Fatalf("Failed to create token for user %s: %v", userID, err)
			}
			tokens[i] = token
		}

		// Validate all tokens
		for i, token := range tokens {
			validatedUserID, err := ValidateJWT(token, tokenSecret)
			if err != nil {
				t.Fatalf("Failed to validate token %d: %v", i, err)
			}

			if validatedUserID != users[i] {
				t.Errorf("Token %d: expected user ID %s, got %s", i, users[i], validatedUserID)
			}
		}
	})
}

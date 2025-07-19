package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	if token == "" {
		return "", errors.New("no authorization header provided")
	}

	parts := strings.Split(token, " ")
	if len(parts) < 2 {
		return "", errors.New("authorization header format must be Bearer {token}")
	}

	return parts[1], nil
}

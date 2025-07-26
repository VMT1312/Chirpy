package auth

import (
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	apiKey := headers.Get("Authorization")
	if apiKey == "" {
		return "", http.ErrNoCookie
	}

	parts := strings.Split(apiKey, " ")
	if len(parts) < 2 || parts[0] != "ApiKey" {
		return "", http.ErrNoCookie
	}
	return parts[1], nil
}

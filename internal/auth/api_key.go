package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorization header is required")
	}

	split := strings.Split(authHeader, " ")
	if len(split) < 2 || split[0] != "ApiKey" {
		return "", errors.New("malformed authorization header")
	}

	return split[1], nil
}

package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}
	split := strings.Split(authHeader, " ")
	if len(split) < 2 || split[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}

	return split[1], nil
}

func MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	stringToken := hex.EncodeToString(token)
	return stringToken, nil
}

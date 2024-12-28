package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	b64 "encoding/base64"
)

type contextKey string

var (
	AuthenticatedUserKey = contextKey("user_id")
)

func validateToken(tokenstring string) (map[string]interface{}, error) {
	url := os.Getenv("TOKEN_VALIDATION_URL")
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("authorization", "Bearer "+tokenstring)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("token not valid %d", resp.StatusCode)
	}
	sections := strings.Split(tokenstring, ".")
	if len(sections) != 3 {
		return nil, fmt.Errorf("bad token")
	}
	decoded := make([]byte, b64.RawStdEncoding.DecodedLen(len(sections[1])))
	_, err = b64.RawStdEncoding.Decode(decoded, []byte(sections[1]))
	if err != nil {
		return nil, err
	}
	claims := make(map[string]interface{})
	err = json.Unmarshal(decoded, &claims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func validateAndGetClaims(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	auth := r.Header.Get("authorization")
	if len(auth) < len("Bearer ") {
		fmt.Printf("Bad Auth\n")
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	auth = auth[7:]
	validClaims, err := validateToken(auth)
	if err != nil {
		fmt.Printf("Error validating Token %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	if validClaims == nil {
		fmt.Printf("Token is not valid\n")
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}

	return validClaims
}

func checkScope(c map[string]interface{}, s string) bool {
	scopesInterface := c["scopes"]
	if scopesInterface == nil {
		return false
	}

	scopes, ok := scopesInterface.(string)
	if !ok {
		return false
	}

	scopeSlice := strings.Split(scopes, " ")
	for is := range scopeSlice {
		if scopeSlice[is] == s {
			return true
		}
	}

	return false
}

type TokenMiddleware struct {
	handler        http.Handler
	required_scope string
}

func (l *TokenMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims := validateAndGetClaims(w, r)
	if claims == nil {
		return
	}
	if !checkScope(claims, l.required_scope) {
		fmt.Printf("Token is not valid no scope.\n")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	uid_iface := claims["user_id"]
	if uid_iface == nil {
		fmt.Printf("Token is not valid no user_id\n")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	uid, ok := uid_iface.(string)
	if !ok || len(uid) <= 0 {
		fmt.Printf("Token is not valid user_id bad\n")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	ctxWithUser := context.WithValue(r.Context(), AuthenticatedUserKey, uid)
	rWithUser := r.WithContext(ctxWithUser)
	l.handler.ServeHTTP(w, rWithUser)
}

func NewTokenMiddleware(handlerToWrap http.Handler, required_scope string) *TokenMiddleware {
	return &TokenMiddleware{handlerToWrap, required_scope}
}

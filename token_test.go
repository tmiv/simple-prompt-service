package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateToken(t *testing.T) {
	// Setup mock token validation server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("authorization")
		if auth == "Bearer eyJhbGciOiJIUzUxMiIsImtpZCI6ImM5YzMzYmEwLWI1YzItNDAyMi1hZjczLTg2NDg2ZDY1OWViMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL2ZsdW8ucGFydHkiLCJleHAiOjE3MzUyNjI1NDksImlhdCI6MTczNTI1MTc0OSwiaWQiOiJiZDg0ZTliZC03ZjcyLTRlOWItYmEyOC0xMjIyZWQ2NmI3MDciLCJpc3MiOiJodHRwczovL2dvLmNvbSIsIm5iZiI6MTczNTI1MTc0OSwic2NvcGVzIjoiaGVsbG8gZ29vZGJ5ZSIsInVzZXJfaWQiOiJxYmVydCJ9.Fp0IEEJ2mX3EH_XW6tRHOb7JffdMbwM7S8MG6h1ke2gcscsFZW9ZJaAHnGuE9ppjMLbZqZd0UMtW4UWjnSmdnw" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer mockServer.Close()

	tokenValidationURL = mockServer.URL

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token format",
			token:   "eyJhbGciOiJIUzUxMiIsImtpZCI6ImM5YzMzYmEwLWI1YzItNDAyMi1hZjczLTg2NDg2ZDY1OWViMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL2ZsdW8ucGFydHkiLCJleHAiOjE3MzUyNjI1NDksImlhdCI6MTczNTI1MTc0OSwiaWQiOiJiZDg0ZTliZC03ZjcyLTRlOWItYmEyOC0xMjIyZWQ2NmI3MDciLCJpc3MiOiJodHRwczovL2dvLmNvbSIsIm5iZiI6MTczNTI1MTc0OSwic2NvcGVzIjoiaGVsbG8gZ29vZGJ5ZSIsInVzZXJfaWQiOiJxYmVydCJ9.Fp0IEEJ2mX3EH_XW6tRHOb7JffdMbwM7S8MG6h1ke2gcscsFZW9ZJaAHnGuE9ppjMLbZqZd0UMtW4UWjnSmdnw",
			wantErr: false,
		},
		{
			name:    "invalid token format",
			token:   "invalid-token",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateToken(tt.token)
			if (err != nil) != tt.wantErr {
				if tt.wantErr {
					t.Errorf("validateToken() wanted error didn't get it")
				} else {
					t.Errorf("validateToken() got unexpected error %v", err)
				}
			}
		})
	}
}

func TestValidateAndGetClaims(t *testing.T) {
	// Setup mock token validation server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("authorization")
		if auth == "Bearer eyJhbGciOiJIUzUxMiIsImtpZCI6ImM5YzMzYmEwLWI1YzItNDAyMi1hZjczLTg2NDg2ZDY1OWViMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL2ZsdW8ucGFydHkiLCJleHAiOjE3MzUyNjI1NDksImlhdCI6MTczNTI1MTc0OSwiaWQiOiJiZDg0ZTliZC03ZjcyLTRlOWItYmEyOC0xMjIyZWQ2NmI3MDciLCJpc3MiOiJodHRwczovL2dvLmNvbSIsIm5iZiI6MTczNTI1MTc0OSwic2NvcGVzIjoiaGVsbG8gZ29vZGJ5ZSIsInVzZXJfaWQiOiJxYmVydCJ9.Fp0IEEJ2mX3EH_XW6tRHOb7JffdMbwM7S8MG6h1ke2gcscsFZW9ZJaAHnGuE9ppjMLbZqZd0UMtW4UWjnSmdnw" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer mockServer.Close()

	tokenValidationURL = mockServer.URL

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "valid auth header",
			authHeader: "Bearer eyJhbGciOiJIUzUxMiIsImtpZCI6ImM5YzMzYmEwLWI1YzItNDAyMi1hZjczLTg2NDg2ZDY1OWViMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL2ZsdW8ucGFydHkiLCJleHAiOjE3MzUyNjI1NDksImlhdCI6MTczNTI1MTc0OSwiaWQiOiJiZDg0ZTliZC03ZjcyLTRlOWItYmEyOC0xMjIyZWQ2NmI3MDciLCJpc3MiOiJodHRwczovL2dvLmNvbSIsIm5iZiI6MTczNTI1MTc0OSwic2NvcGVzIjoiaGVsbG8gZ29vZGJ5ZSIsInVzZXJfaWQiOiJxYmVydCJ9.Fp0IEEJ2mX3EH_XW6tRHOb7JffdMbwM7S8MG6h1ke2gcscsFZW9ZJaAHnGuE9ppjMLbZqZd0UMtW4UWjnSmdnw",
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing Bearer prefix",
			authHeader: "eyJhbGciOiJIUzUxMiIsImtpZCI6ImM5YzMzYmEwLWI1YzItNDAyMi1hZjczLTg2NDg2ZDY1OWViMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL2ZsdW8ucGFydHkiLCJleHAiOjE3MzUyNjI1NDksImlhdCI6MTczNTI1MTc0OSwiaWQiOiJiZDg0ZTliZC03ZjcyLTRlOWItYmEyOC0xMjIyZWQ2NmI3MDciLCJpc3MiOiJodHRwczovL2dvLmNvbSIsIm5iZiI6MTczNTI1MTc0OSwic2NvcGVzIjoiaGVsbG8gZ29vZGJ5ZSIsInVzZXJfaWQiOiJxYmVydCJ9.Fp0IEEJ2mX3EH_XW6tRHOb7JffdMbwM7S8MG6h1ke2gcscsFZW9ZJaAHnGuE9ppjMLbZqZd0UMtW4UWjnSmdnw",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty auth header",
			authHeader: "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			claims := validateAndGetClaims(w, req)
			if tt.wantStatus != http.StatusOK && claims != nil {
				t.Error("expected nil claims for invalid request")
			}
			if w.Code != tt.wantStatus {
				t.Errorf("validateAndGetClaims() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestCheckScope(t *testing.T) {
	tests := []struct {
		name   string
		claims map[string]interface{}
		scope  string
		want   bool
	}{
		{
			name: "valid scope present",
			claims: map[string]interface{}{
				"scopes": "read write delete",
			},
			scope: "write",
			want:  true,
		},
		{
			name: "scope not present",
			claims: map[string]interface{}{
				"scopes": "read delete",
			},
			scope: "write",
			want:  false,
		},
		{
			name:   "nil claims",
			claims: nil,
			scope:  "write",
			want:   false,
		},
		{
			name: "invalid scopes type",
			claims: map[string]interface{}{
				"scopes": 123,
			},
			scope: "write",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkScope(tt.claims, tt.scope); got != tt.want {
				t.Errorf("checkScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenMiddleware(t *testing.T) {
	// Setup mock token validation server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	tokenValidationURL = mockServer.URL

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		authHeader string
		scope      string
		wantStatus int
	}{
		{
			name:       "valid request",
			authHeader: "Bearer eyJhbGciOiJIUzUxMiIsImtpZCI6ImM5YzMzYmEwLWI1YzItNDAyMi1hZjczLTg2NDg2ZDY1OWViMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL2ZsdW8ucGFydHkiLCJleHAiOjE3MzUyNjI1NDksImlhdCI6MTczNTI1MTc0OSwiaWQiOiJiZDg0ZTliZC03ZjcyLTRlOWItYmEyOC0xMjIyZWQ2NmI3MDciLCJpc3MiOiJodHRwczovL2dvLmNvbSIsIm5iZiI6MTczNTI1MTc0OSwic2NvcGVzIjoiaGVsbG8gZ29vZGJ5ZSIsInVzZXJfaWQiOiJxYmVydCJ9.Fp0IEEJ2mX3EH_XW6tRHOb7JffdMbwM7S8MG6h1ke2gcscsFZW9ZJaAHnGuE9ppjMLbZqZd0UMtW4UWjnSmdnw",
			scope:      "read",
			wantStatus: http.StatusUnauthorized, // Will be unauthorized because our mock token doesn't have valid claims
		},
		{
			name:       "missing auth header",
			authHeader: "",
			scope:      "read",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "has goodbye scope",
			authHeader: "Bearer eyJhbGciOiJIUzUxMiIsImtpZCI6ImM5YzMzYmEwLWI1YzItNDAyMi1hZjczLTg2NDg2ZDY1OWViMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL2ZsdW8ucGFydHkiLCJleHAiOjE3MzUyNjI1NDksImlhdCI6MTczNTI1MTc0OSwiaWQiOiJiZDg0ZTliZC03ZjcyLTRlOWItYmEyOC0xMjIyZWQ2NmI3MDciLCJpc3MiOiJodHRwczovL2dvLmNvbSIsIm5iZiI6MTczNTI1MTc0OSwic2NvcGVzIjoiaGVsbG8gZ29vZGJ5ZSIsInVzZXJfaWQiOiJxYmVydCJ9.Fp0IEEJ2mX3EH_XW6tRHOb7JffdMbwM7S8MG6h1ke2gcscsFZW9ZJaAHnGuE9ppjMLbZqZd0UMtW4UWjnSmdnw",
			scope:      "goodbye",
			wantStatus: http.StatusOK,
		},
		{
			name:       "has seeya scope",
			authHeader: "Bearer eyJhbGciOiJIUzUxMiIsImtpZCI6ImM5YzMzYmEwLWI1YzItNDAyMi1hZjczLTg2NDg2ZDY1OWViMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL2ZsdW8ucGFydHkiLCJleHAiOjE3MzUyNjI1NDksImlhdCI6MTczNTI1MTc0OSwiaWQiOiJiZDg0ZTliZC03ZjcyLTRlOWItYmEyOC0xMjIyZWQ2NmI3MDciLCJpc3MiOiJodHRwczovL2dvLmNvbSIsIm5iZiI6MTczNTI1MTc0OSwic2NvcGVzIjoiaGVsbG8gZ29vZGJ5ZSIsInVzZXJfaWQiOiJxYmVydCJ9.Fp0IEEJ2mX3EH_XW6tRHOb7JffdMbwM7S8MG6h1ke2gcscsFZW9ZJaAHnGuE9ppjMLbZqZd0UMtW4UWjnSmdnw",
			scope:      "seeya",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "has hello scope",
			authHeader: "Bearer eyJhbGciOiJIUzUxMiIsImtpZCI6ImM5YzMzYmEwLWI1YzItNDAyMi1hZjczLTg2NDg2ZDY1OWViMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL2ZsdW8ucGFydHkiLCJleHAiOjE3MzUyNjI1NDksImlhdCI6MTczNTI1MTc0OSwiaWQiOiJiZDg0ZTliZC03ZjcyLTRlOWItYmEyOC0xMjIyZWQ2NmI3MDciLCJpc3MiOiJodHRwczovL2dvLmNvbSIsIm5iZiI6MTczNTI1MTc0OSwic2NvcGVzIjoiaGVsbG8gZ29vZGJ5ZSIsInVzZXJfaWQiOiJxYmVydCJ9.Fp0IEEJ2mX3EH_XW6tRHOb7JffdMbwM7S8MG6h1ke2gcscsFZW9ZJaAHnGuE9ppjMLbZqZd0UMtW4UWjnSmdnw",
			scope:      "hello",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewTokenMiddleware(testHandler, tt.scope)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("TokenMiddleware.ServeHTTP() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

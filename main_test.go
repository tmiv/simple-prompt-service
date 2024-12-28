package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Save original env var if it exists
	originalPrompts := os.Getenv("PROMPTS")

	// Set test environment variable
	testPrompts := `{
		"test": {
			"service": "anthropic",
			"model": "claude-3",
			"system": "You are a helpful assistant",
			"maxTokens": 1000,
			"cost": {
				"path": "test/path",
				"cost": 1
			}
		}
	}`
	os.Setenv("PROMPTS", testPrompts)

	// Run tests
	code := m.Run()

	// Restore original env var
	if originalPrompts != "" {
		os.Setenv("PROMPTS", originalPrompts)
	} else {
		os.Unsetenv("PROMPTS")
	}

	os.Exit(code)
}

func TestSetupCors(t *testing.T) {
	tests := []struct {
		name        string
		corsOrigins string
		wantDefault bool
	}{
		{
			name:        "with origins",
			corsOrigins: "http://localhost:3000,http://example.com",
			wantDefault: false,
		},
		{
			name:        "empty origins",
			corsOrigins: "",
			wantDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env var
			originalCorsOrigins := os.Getenv("CORS_ORIGINS")

			// Set test env var
			os.Setenv("CORS_ORIGINS", tt.corsOrigins)

			cors := setupcors()
			if cors == nil {
				t.Error("setupcors() returned nil")
			}

			// Restore original env var
			if originalCorsOrigins != "" {
				os.Setenv("CORS_ORIGINS", originalCorsOrigins)
			} else {
				os.Unsetenv("CORS_ORIGINS")
			}
		})
	}
}

func TestConstructPromptHandler(t *testing.T) {
	// Create a test prompt declaration
	testPrompt := PromptDeclaration{
		Service:   Anthropic,
		Model:     "claude-3",
		MaxTokens: 1000,
	}

	// Create the handler
	handler := constructPromptHandler(&testPrompt)

	// TODO: Add more specific tests for the handler functionality
	if handler == nil {
		t.Error("constructPromptHandler() returned nil")
	}
}

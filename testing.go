//go:build test
// +build test

package main

import "os"

// SetupTestEnvironment configures the environment for testing
func SetupTestEnvironment() func() {
	// Save original env vars
	originalPrompts := os.Getenv("PROMPTS")

	// Set test environment variables
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

	// Return cleanup function
	return func() {
		if originalPrompts != "" {
			os.Setenv("PROMPTS", originalPrompts)
		} else {
			os.Unsetenv("PROMPTS")
		}
	}
}

package main

import (
	"testing"

	fcs "github.com/tmiv/firebase-credit-service"
)

func TestValidatePromptDeclaration(t *testing.T) {
	// Helper function to create a string pointer
	strPtr := func(s string) *string {
		return &s
	}

	// Create a valid prompt declaration for reuse
	validPrompt := PromptDeclaration{
		Service:   Anthropic,
		Model:     "claude-3",
		System:    strPtr("You are a helpful assistant"),
		MaxTokens: 1000,
		Cost: fcs.ChargeData{
			Path: "test/path",
			Cost: 1,
		},
		Temperature:   0,
		Variables:     []string{},
		RequiredScope: "hello",
	}

	tests := []struct {
		name    string
		prompt  *PromptDeclaration
		wantErr bool
	}{
		{
			name:    "nil prompt",
			prompt:  nil,
			wantErr: true,
		},
		{
			name:    "valid prompt",
			prompt:  &validPrompt,
			wantErr: false,
		},
		{
			name: "invalid service",
			prompt: &PromptDeclaration{
				Service:   "invalid",
				Model:     "claude-3",
				System:    strPtr("test"),
				MaxTokens: 1000,
				Cost: fcs.ChargeData{
					Path: "test/path",
					Cost: 1,
				},
				Temperature:   0,
				Variables:     []string{},
				RequiredScope: "hello",
			},
			wantErr: true,
		},
		{
			name: "empty model",
			prompt: &PromptDeclaration{
				Service:   Anthropic,
				Model:     "",
				System:    strPtr("test"),
				MaxTokens: 1000,
				Cost: fcs.ChargeData{
					Path: "test/path",
					Cost: 1,
				},
				Temperature:   0,
				Variables:     []string{},
				RequiredScope: "hello",
			},
			wantErr: true,
		},
		{
			name: "invalid max tokens",
			prompt: &PromptDeclaration{
				Service:   Anthropic,
				Model:     "claude-3",
				System:    strPtr("test"),
				MaxTokens: 0,
				Cost: fcs.ChargeData{
					Path: "test/path",
					Cost: 1,
				},
				Temperature:   0,
				Variables:     []string{},
				RequiredScope: "hello",
			},
			wantErr: true,
		},
		{
			name: "missing cost path",
			prompt: &PromptDeclaration{
				Service:   Anthropic,
				Model:     "claude-3",
				System:    strPtr("test"),
				MaxTokens: 1000,
				Cost: fcs.ChargeData{
					Path: "",
					Cost: 1,
				},
				Temperature:   0,
				Variables:     []string{},
				RequiredScope: "hello",
			},
			wantErr: true,
		},
		{
			name: "negative cost",
			prompt: &PromptDeclaration{
				Service:   Anthropic,
				Model:     "claude-3",
				System:    strPtr("test"),
				MaxTokens: 1000,
				Cost: fcs.ChargeData{
					Path: "test/path",
					Cost: -1,
				},
				Temperature:   0,
				Variables:     []string{},
				RequiredScope: "hello",
			},
			wantErr: true,
		},
		{
			name: "missing system and initial user",
			prompt: &PromptDeclaration{
				Service:   Anthropic,
				Model:     "claude-3",
				MaxTokens: 1000,
				Cost: fcs.ChargeData{
					Path: "test/path",
					Cost: 1,
				},
				Temperature:   0,
				Variables:     []string{},
				RequiredScope: "hello",
			},
			wantErr: true,
		},
		{
			name: "valid with initial user instead of system",
			prompt: &PromptDeclaration{
				Service:     Anthropic,
				Model:       "claude-3",
				InitialUser: strPtr("Hello, AI"),
				MaxTokens:   1000,
				Cost: fcs.ChargeData{
					Path: "test/path",
					Cost: 1,
				},
				Temperature:   0,
				Variables:     []string{},
				RequiredScope: "hello",
			},
			wantErr: false,
		},
		{
			name: "missing required scope",
			prompt: &PromptDeclaration{
				Service:     Anthropic,
				Model:       "claude-3",
				InitialUser: strPtr("Hello, AI"),
				MaxTokens:   1000,
				Cost: fcs.ChargeData{
					Path: "test/path",
					Cost: 1,
				},
				Temperature: 0,
				Variables:   []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePromptDeclartion(tt.name, tt.prompt)
			if result == tt.wantErr {
				t.Errorf("ValidatePromptDeclartion() = %v, want %v", result, !tt.wantErr)
			}
		})
	}
}

func TestValidatePromptConfig(t *testing.T) {
	strPtr := func(s string) *string {
		return &s
	}

	validPrompt := PromptDeclaration{
		Service:   Anthropic,
		Model:     "claude-3",
		System:    strPtr("You are a helpful assistant"),
		MaxTokens: 1000,
		Cost: fcs.ChargeData{
			Path: "test/path",
			Cost: 1,
		},
		Temperature:   0,
		Variables:     []string{},
		RequiredScope: "hello",
	}

	invalidPrompt := PromptDeclaration{
		Service:   "invalid",
		Model:     "claude-3",
		System:    strPtr("test"),
		MaxTokens: 1000,
		Cost: fcs.ChargeData{
			Path: "test/path",
			Cost: 1,
		},
		Temperature:   0,
		Variables:     []string{},
		RequiredScope: "hello",
	}

	tests := []struct {
		name   string
		config PromptConfig
		want   bool
	}{
		{
			name: "valid config",
			config: PromptConfig{
				"test1": validPrompt,
				"test2": validPrompt,
			},
			want: true,
		},
		{
			name: "invalid config",
			config: PromptConfig{
				"test1": validPrompt,
				"test2": invalidPrompt,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidatePromptConfig(&tt.config); got != tt.want {
				t.Errorf("ValidatePromptConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

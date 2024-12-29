package main

import (
	"net/http"
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

func TestMakeResult(t *testing.T) {
	tests := []struct {
		name    string
		context []byte
		result  string
		wantErr bool
	}{
		{
			name:    "valid result",
			context: []byte("test context"),
			result:  "test result",
			wantErr: false,
		},
		{
			name:    "empty context",
			context: []byte{},
			result:  "test result",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeResult(tt.context, tt.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Error("MakeResult() returned nil response")
					return
				}
				if got.Result != tt.result {
					t.Errorf("MakeResult() result = %v, want %v", got.Result, tt.result)
				}
				if got.Context == "" {
					t.Error("MakeResult() returned empty context")
				}
			}
		})
	}
}

func TestCollectVariables(t *testing.T) {
	tests := []struct {
		name       string
		variables  []string
		formValues map[string]string
		wantVars   PromptVariables
	}{
		{
			name:      "collect single variable",
			variables: []string{"name"},
			formValues: map[string]string{
				"name": "test",
			},
			wantVars: PromptVariables{
				"name": "test",
			},
		},
		{
			name:      "collect multiple variables",
			variables: []string{"name", "age"},
			formValues: map[string]string{
				"name": "test",
				"age":  "25",
			},
			wantVars: PromptVariables{
				"name": "test",
				"age":  "25",
			},
		},
		{
			name:       "missing form value",
			variables:  []string{"name"},
			formValues: map[string]string{},
			wantVars: PromptVariables{
				"name": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := createTestRequest(tt.formValues)
			p := &PromptDeclaration{Variables: tt.variables}

			got := CollectVariables(r, p)

			if len(got) != len(tt.wantVars) {
				t.Errorf("CollectVariables() got %v vars, want %v", len(got), len(tt.wantVars))
			}

			for k, v := range tt.wantVars {
				if got[k] != v {
					t.Errorf("CollectVariables() got[%s] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestCollectContinuanceVariables(t *testing.T) {
	tests := []struct {
		name       string
		formValues map[string]string
		want       PromptVariables
	}{
		{
			name: "collect both variables",
			formValues: map[string]string{
				"USER_TEXT": "hello",
				"CONTEXT":   "previous context",
			},
			want: PromptVariables{
				"USER_TEXT": "hello",
				"CONTEXT":   "previous context",
			},
		},
		{
			name:       "missing values",
			formValues: map[string]string{},
			want: PromptVariables{
				"USER_TEXT": "",
				"CONTEXT":   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := createTestRequest(tt.formValues)

			got := CollectContinuanceVariables(r)

			if len(got) != len(tt.want) {
				t.Errorf("CollectContinuanceVariables() got %v vars, want %v", len(got), len(tt.want))
			}

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("CollectContinuanceVariables() got[%s] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

// Helper function to create test HTTP requests
func createTestRequest(formValues map[string]string) *http.Request {
	r, _ := http.NewRequest("POST", "/", nil)
	r.PostForm = make(map[string][]string)
	for k, v := range formValues {
		r.PostForm[k] = []string{v}
	}
	return r
}

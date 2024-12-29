package main

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectLatestResponses(t *testing.T) {
	tests := []struct {
		name     string
		messages []messageParam
		want     string
	}{
		{
			name: "single assistant response",
			messages: []messageParam{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there"},
			},
			want: "Hi there",
		},
		{
			name: "multiple assistant responses",
			messages: []messageParam{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there"},
				{Role: "user", Content: "How are you?"},
				{Role: "assistant", Content: "I'm good"},
			},
			want: "I'm good",
		},
		{
			name: "consecutive assistant responses",
			messages: []messageParam{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Response 1"},
				{Role: "assistant", Content: "Response 2"},
			},
			want: "Response 1\nResponse 2",
		},
		{
			name:     "empty messages",
			messages: []messageParam{},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &anthropicRequest{Messages: tt.messages}
			got := collectLatestResponses(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildRequest(t *testing.T) {
	tests := []struct {
		name     string
		prompt   *PromptDeclaration
		vars     PromptVariables
		wantErr  bool
		validate func(*testing.T, *anthropicRequest)
	}{
		{
			name: "basic request",
			prompt: &PromptDeclaration{
				Model:       "claude-3-sonnet",
				MaxTokens:   1000,
				Temperature: 0.7,
				System:      stringPtr("System prompt {{VAR}}"),
				InitialUser: stringPtr("User message {{VAR}}"),
			},
			vars: PromptVariables{"VAR": "test"},
			validate: func(t *testing.T, req *anthropicRequest) {
				assert.Equal(t, "claude-3-sonnet", req.Model)
				assert.Equal(t, 1000, req.MaxTokens)
				assert.Equal(t, float32(0.7), req.Temperature)
				assert.Equal(t, "System prompt test", *req.System)
				assert.Equal(t, 1, len(req.Messages))
				assert.Equal(t, "User message test", req.Messages[0].Content)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _, err := buildRequest(tt.prompt, tt.vars)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			tt.validate(t, req)
		})
	}
}

func TestBuildContinueRequest(t *testing.T) {
	tests := []struct {
		name    string
		context string
		text    string
		wantErr bool
	}{
		{
			name: "valid continue request",
			context: `{
				"model": "claude-3-sonnet",
				"messages": [
					{"role": "user", "content": "Hello"}
				]
			}`,
			text: "How are you?",
		},
		{
			name:    "missing context",
			context: "",
			text:    "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := PromptVariables{
				"CONTEXT":   tt.context,
				"USER_TEXT": tt.text,
			}
			req, _, err := buildContinueRequest(&PromptDeclaration{}, vars)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.text, req.Messages[len(req.Messages)-1].Content)
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
func TestBackfillResponse(t *testing.T) {
	tests := []struct {
		name     string
		request  *anthropicRequest
		response anthropicResponse
		want     []messageParam
	}{
		{
			name: "basic backfill",
			request: &anthropicRequest{
				Messages: []messageParam{
					{Role: "user", Content: "Hello"},
				},
			},
			response: anthropicResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "text", Text: "Hi there"},
				},
			},
			want: []messageParam{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there"},
			},
		},
		{
			name: "multiple content types",
			request: &anthropicRequest{
				Messages: []messageParam{
					{Role: "user", Content: "Hello"},
				},
			},
			response: anthropicResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "other", Text: "ignored"},
					{Type: "text", Text: "Hi there"},
				},
			},
			want: []messageParam{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := backfillReponse(tt.request, tt.response)
			assert.Equal(t, tt.want, result.Messages)
		})
	}
}

func TestPackageResult(t *testing.T) {
	tests := []struct {
		name       string
		respBody   string
		wantErr    bool
		wantResult string
	}{
		{
			name: "successful response",
			respBody: `{
                "content": [{"type": "text", "text": "Hello!"}],
                "role": "assistant"
            }`,
			wantResult: "Hello!",
		},
		{
			name: "error response",
			respBody: `{
                "error": {"message": "Invalid request"}
            }`,
			wantErr: true,
		},
		{
			name: "empty content",
			respBody: `{
                "content": []
            }`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Body: io.NopCloser(strings.NewReader(tt.respBody)),
			}
			reqBody := &anthropicRequest{
				Messages: []messageParam{
					{Role: "user", Content: "Hi"},
				},
			}

			_, result, err := packageResult(resp, reqBody)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}

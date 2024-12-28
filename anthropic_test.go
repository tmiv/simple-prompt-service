package main

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestBackfillResponse(t *testing.T) {
	req := anthropicRequest{
		Model:     "claude-3-opus-20240229",
		Messages:  []messageParam{},
		MaxTokens: 1024,
	}

	resp := anthropicResponse{
		Content: []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{
			{Type: "text", Text: "Hello world"},
		},
	}

	result := backfillReponse(&req, resp)

	if len(result.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(result.Messages))
	}

	if result.Messages[0].Role != "assistant" {
		t.Errorf("Expected role 'assistant', got %s", result.Messages[0].Role)
	}

	if result.Messages[0].Content != "Hello world" {
		t.Errorf("Expected content 'Hello world', got %s", result.Messages[0].Content)
	}
}

func TestCollectLatestResponses(t *testing.T) {
	tests := []struct {
		name     string
		messages []messageParam
		want     string
	}{
		{
			name: "multiple assistant responses",
			messages: []messageParam{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi"},
				{Role: "assistant", Content: "How are you?"},
			},
			want: "Hi\nHow are you?",
		},
		{
			name: "single assistant response",
			messages: []messageParam{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi"},
			},
			want: "Hi",
		},
		{
			name: "no assistant responses",
			messages: []messageParam{
				{Role: "user", Content: "Hello"},
			},
			want: "",
		},
		{
			name: "mixed responses with user interruption",
			messages: []messageParam{
				{Role: "assistant", Content: "First"},
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Second"},
				{Role: "assistant", Content: "Third"},
			},
			want: "Second\nThird",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := anthropicRequest{
				Messages: tt.messages,
			}

			got := collectLatestResponses(&req)
			if got != tt.want {
				t.Errorf("collectLatestResponses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildRequest(t *testing.T) {
	form := url.Values{}
	form.Add("name", "Test User")
	form.Add("topic", "Testing")

	req, _ := http.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()

	systemPrompt := "You are talking to {{name}} about {{topic}}"
	userPrompt := "Hello {{name}}, let's discuss {{topic}}"
	agentPrompt := "I'll help {{name}} with {{topic}}"

	tests := []struct {
		name    string
		prompt  *PromptDeclaration
		want    *anthropicRequest
		wantErr bool
	}{
		{
			name: "full request with all fields",
			prompt: &PromptDeclaration{
				Model:        "claude-3-opus-20240229",
				MaxTokens:    1024,
				Temperature:  0.7,
				System:       &systemPrompt,
				InitialUser:  &userPrompt,
				InitialAgent: &agentPrompt,
				Variables:    []string{"name", "topic"},
			},
			want: &anthropicRequest{
				Model:       "claude-3-opus-20240229",
				MaxTokens:   1024,
				Temperature: 0.7,
				System:      stringPtr("You are talking to Test User about Testing"),
				Messages: []messageParam{
					{Role: "user", Content: "Hello Test User, let's discuss Testing"},
					{Role: "assistant", Content: "I'll help Test User with Testing"},
				},
			},
			wantErr: false,
		},
		{
			name: "request without system prompt",
			prompt: &PromptDeclaration{
				Model:       "claude-3-opus-20240229",
				MaxTokens:   1024,
				Temperature: 0.7,
				InitialUser: &userPrompt,
				Variables:   []string{"name", "topic"},
			},
			want: &anthropicRequest{
				Model:       "claude-3-opus-20240229",
				MaxTokens:   1024,
				Temperature: 0.7,
				Messages: []messageParam{
					{Role: "user", Content: "Hello Test User, let's discuss Testing"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := buildRequest(req, tt.prompt)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPackageResult(t *testing.T) {
	tests := []struct {
		name       string
		response   *http.Response
		reqBody    *anthropicRequest
		wantResult string
		wantErr    bool
	}{
		{
			name: "successful response",
			response: &http.Response{
				Body: io.NopCloser(strings.NewReader(`{
					"content": [{"type": "text", "text": "Hello world"}],
					"role": "assistant",
					"model": "claude-3-opus-20240229"
				}`)),
			},
			reqBody: &anthropicRequest{
				Model:    "claude-3-opus-20240229",
				Messages: []messageParam{},
			},
			wantResult: "Hello world",
			wantErr:    false,
		},
		{
			name: "error in response",
			response: &http.Response{
				Body: io.NopCloser(strings.NewReader(`{
					"error": {"message": "Invalid API key"}
				}`)),
			},
			reqBody:    &anthropicRequest{},
			wantResult: "",
			wantErr:    true,
		},
		{
			name: "empty content",
			response: &http.Response{
				Body: io.NopCloser(strings.NewReader(`{
					"content": [],
					"role": "assistant",
					"model": "claude-3-opus-20240229"
				}`)),
			},
			reqBody:    &anthropicRequest{},
			wantResult: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := packageResult(tt.response, tt.reqBody)
			if (err != nil) != tt.wantErr {
				t.Errorf("packageResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.wantResult {
				t.Errorf("packageResult() result = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

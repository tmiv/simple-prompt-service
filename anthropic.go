package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var (
	anthropicMessageEndpoint = "https://api.anthropic.com/v1/messages"
	anthropicVersion         = "2023-06-01"
)

type messageParam struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type anthropicRequest struct {
	Model       string         `json:"model"`
	MaxTokens   int            `json:"max_tokens"`
	System      *string        `json:"system,omitempty"`
	Temperature float32        `json:"temperature"`
	Messages    []messageParam `json:"messages"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string  `json:"model"`
	StopReason   string  `json:"stop_reason"`
	StopSequence *string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type variables map[string]string

func backfillReponse(req *anthropicRequest, resp anthropicResponse) *anthropicRequest {
	var responseText string
	for _, content := range resp.Content {
		if content.Type == "text" {
			responseText = content.Text
			break
		}
	}

	req.Messages = append(req.Messages, messageParam{
		Role:    "assistant",
		Content: responseText,
	})

	return req
}

func collectLatestResponses(req *anthropicRequest) string {
	var responses []string

	// Iterate through messages in reverse order
	for i := len(req.Messages) - 1; i >= 0; i-- {
		msg := req.Messages[i]
		if msg.Role != "assistant" {
			break
		}
		responses = append(responses, msg.Content)
	}

	// Reverse the responses to maintain chronological order
	for i := 0; i < len(responses)/2; i++ {
		j := len(responses) - 1 - i
		responses[i], responses[j] = responses[j], responses[i]
	}

	// Join all responses with newlines
	return strings.Join(responses, "\n")
}

func buildRequest(r *http.Request, p *PromptDeclaration) (*anthropicRequest, []byte, error) {
	reqBody := anthropicRequest{
		Model:       p.Model,
		MaxTokens:   p.MaxTokens,
		Temperature: p.Temperature,
		Messages:    []messageParam{},
	}

	vars := make(variables)
	for key := range p.Variables {
		varKey := p.Variables[key]
		vars[varKey] = r.FormValue(varKey)
	}

	// Apply variables to system prompt if present
	if p.System != nil {
		systemPrompt := *p.System
		for key, value := range vars {
			systemPrompt = strings.ReplaceAll(systemPrompt, fmt.Sprintf("{{%s}}", key), value)
		}
		reqBody.System = &systemPrompt
	}

	// Apply variables to initial messages if present
	if p.InitialUser != nil {
		userPrompt := *p.InitialUser
		for key, value := range vars {
			userPrompt = strings.ReplaceAll(userPrompt, fmt.Sprintf("{{%s}}", key), value)
		}
		reqBody.Messages = append(reqBody.Messages, messageParam{
			Role:    "user",
			Content: userPrompt,
		})
	}
	if p.InitialAgent != nil {
		agentPrompt := *p.InitialAgent
		for key, value := range vars {
			agentPrompt = strings.ReplaceAll(agentPrompt, fmt.Sprintf("{{%s}}", key), value)
		}
		reqBody.Messages = append(reqBody.Messages, messageParam{
			Role:    "assistant",
			Content: agentPrompt,
		})
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling request: %w", err)
	}

	return &reqBody, jsonBody, nil
}

func AnthropicProcessPrompt(r *http.Request, p *PromptDeclaration) ([]byte, string, error) {
	client := &http.Client{}

	reqBody, jsonBody, err := buildRequest(r, p)
	if err != nil {
		return nil, "", fmt.Errorf("error creating request content: %w", err)
	}

	req, err := http.NewRequest("POST", anthropicMessageEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("x-api-key", os.Getenv("ANTHROPIC_API_KEY"))
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	result, contBytes, err := packageResult(resp, reqBody)
	if err != nil {
		return nil, "", err
	}

	return contBytes, result, nil
}

func packageResult(resp *http.Response, reqBody *anthropicRequest) (string, []byte, error) {
	var anthResponse anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthResponse); err != nil {
		return "", nil, fmt.Errorf("error decoding response: %w", err)
	}

	if anthResponse.Error != nil {
		return "", nil, fmt.Errorf("API error: %s", anthResponse.Error.Message)
	}

	if len(anthResponse.Content) == 0 {
		return "", nil, fmt.Errorf("no content in response")
	}

	cont := backfillReponse(reqBody, anthResponse)
	result := collectLatestResponses(cont)

	contBytes, err := json.Marshal(cont)
	if err != nil {
		return "", nil, fmt.Errorf("error encoding response: %w", err)
	}
	return result, contBytes, nil
}

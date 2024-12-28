package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"

	fcs "github.com/tmiv/firebase-credit-service"
)

type ServiceType string

const (
	Anthropic ServiceType = "anthropic"
	OpenAI    ServiceType = "openai"
	Gemini    ServiceType = "gemini"
)

type PromptDeclaration struct {
	Service       ServiceType    `json:"service"` // 'anthropic', 'openai', or 'gemini'
	Model         string         `json:"model"`
	System        *string        `json:"system,omitempty"`
	MaxTokens     int            `json:"max_tokens"`
	Temperature   float32        `json:"temperature"`
	InitialUser   *string        `json:"initial_user,omitempty"`
	InitialAgent  *string        `json:"initial_agent,omitempty"`
	Cost          fcs.ChargeData `json:"cost"`
	RequiredScope string         `json:"required_scope"`
	Variables     []string       `json:"variables,omitempty"`
}

type Response struct {
	Context string `json:"context"`
	Result  string `json:"result"`
}

type PromptConfig map[string]PromptDeclaration

func ValidatePromptDeclartion(name string, pd *PromptDeclaration) bool {
	// Check if pointer is nil
	if pd == nil {
		fmt.Printf("Prompt required for %s\n", name)
		return false
	}

	// Validate Service field
	switch pd.Service {
	case Anthropic, OpenAI, Gemini:
		// Valid service type
	default:
		fmt.Printf("service required for %s\n", name)
		return false
	}

	if pd.Model == "" {
		fmt.Printf("model required for %s\n", name)
		return false
	}

	if pd.MaxTokens <= 0 {
		fmt.Printf("max_tokens required for %s\n", name)
		return false
	}

	if pd.Temperature < 0 || pd.Temperature > 1 {
		fmt.Printf("Temperature required for %s\n", name)
		return false
	}

	if pd.Cost.Path == "" {
		fmt.Printf("cost.path required for %s\n", name)
		return false
	}

	if pd.Cost.Cost < 0 {
		fmt.Printf("cost.cost required for %s\n", name)
		return false
	}

	if pd.System == nil && pd.InitialUser == nil {
		fmt.Printf("system or initial user required for %s\n", name)
		return false
	}

	if pd.RequiredScope == "" {
		fmt.Printf("required_scope required for %s\n", name)
		return false
	}

	return true
}

func ValidatePromptConfig(pc *PromptConfig) bool {
	for k, v := range *pc {
		if !ValidatePromptDeclartion(k, &v) {
			return false
		}
	}
	return true
}

func MakeResult(c []byte, r string) (*Response, error) {
	var buf bytes.Buffer

	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}

	_, err = zw.Write(c)
	if err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	compressed := buf.Bytes()
	encrypted, err := Encrypt(compressed)
	if err != nil {
		return nil, err
	}

	return &Response{
		Context: base64.StdEncoding.EncodeToString(encrypted),
		Result:  r,
	}, nil
}

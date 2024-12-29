package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	fcs "github.com/tmiv/firebase-credit-service"
)

type ServiceType string

const (
	Anthropic ServiceType = "anthropic"
	OpenAI    ServiceType = "openai"
	Gemini    ServiceType = "gemini"
)

type PromptVariables map[string]string
type ModelExecutor func(p *PromptDeclaration, vars PromptVariables) (interface{}, string, error)

type PromptDeclaration struct {
	Service            ServiceType     `json:"service"` // 'anthropic', 'openai', or 'gemini'
	Model              string          `json:"model"`
	System             *string         `json:"system,omitempty"`
	MaxTokens          int             `json:"max_tokens"`
	Temperature        float32         `json:"temperature"`
	InitialUser        *string         `json:"initial_user,omitempty"`
	InitialAgent       *string         `json:"initial_agent,omitempty"`
	Cost               fcs.ChargeData  `json:"cost"`
	ContinueCost       *fcs.ChargeData `json:"continue_cost,omitempty"`
	RequiredScope      string          `json:"required_scope"`
	Variables          []string        `json:"variables,omitempty"`
	InitialCreditGrant int             `json:"initial_credit_grant"`
}

type Response struct {
	Context string `json:"context"`
	Result  string `json:"result"`
}

type PromptContext struct {
	Prompt       string      `json:"prompt"`
	ModelContext interface{} `json:"model_context"`
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

func UnpackContext(contextb64 string) (string, string, error) {
	contextCompress, err := base64.StdEncoding.DecodeString(contextb64)
	if err != nil {
		return "", "", fmt.Errorf("decoding CONTEXT %v", err)
	}

	decrypted, err := Decrypt(contextCompress)
	if err != nil {
		return "", "", fmt.Errorf("decrypting context: %v", err)
	}

	zr, err := gzip.NewReader(bytes.NewReader(decrypted))
	if err != nil {
		return "", "", fmt.Errorf("creating gzip reader: %v", err)
	}
	defer zr.Close()

	var pc PromptContext
	if err := json.NewDecoder(zr).Decode(&pc); err != nil {
		return "", "", fmt.Errorf("decoding context JSON: %v", err)
	}

	context, err := json.Marshal(pc.ModelContext)
	if err != nil {
		return "", "", fmt.Errorf("unpacking model context: %v", err)
	}
	return pc.Prompt, string(context), nil
}

func CollectVariables(r *http.Request, p *PromptDeclaration) PromptVariables {
	vars := make(PromptVariables)
	for key := range p.Variables {
		varKey := p.Variables[key]
		vars[varKey] = r.FormValue(varKey)
	}
	return vars
}

func CollectContinuanceVariables(r *http.Request, context string) PromptVariables {
	vars := make(PromptVariables)
	vars["USER_TEXT"] = r.FormValue("USER_TEXT")
	vars["CONTEXT"] = context
	return vars
}

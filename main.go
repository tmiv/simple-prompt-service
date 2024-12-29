package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
	fcs "github.com/tmiv/firebase-credit-service"
)

var (
	prompts     PromptConfig
	firebaseURL string
)

func init() {
	if os.Getenv("TESTING") == "true" {
		return
	}

	promptsJson := os.Getenv("PROMPTS")
	if promptsJson == "" {
		fmt.Println("Fatal: PROMPTS environment variable is not set")
		os.Exit(1)
	}

	if err := json.Unmarshal([]byte(promptsJson), &prompts); err != nil {
		fmt.Printf("Fatal: Failed to parse PROMPTS environment variable: %v\n", err)
		os.Exit(1)
	}

	firebaseURL = os.Getenv("FIREBASE_DB_URL")
	if len(firebaseURL) < 1 {
		fmt.Printf("FIREBASE_DB_URL is not defined")
		os.Exit(1)
	}

	if !ValidatePromptConfig(&prompts) {
		fmt.Printf("Fatal: PROMPTS environment variable invalid\n")
		os.Exit(1)
	}
}

func constructPromptHandler(name string, p *PromptDeclaration) http.HandlerFunc {
	creditService := fcs.NewService(p.Cost, firebaseURL)
	return func(w http.ResponseWriter, r *http.Request) {
		if p.Service != Anthropic {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		vars := CollectVariables(r, p)
		executor := AnthropicProcessPrompt
		runFunc(r.Context(), creditService, name, p, vars, executor, w)
	}
}

func runFunc(ctx context.Context, creditService *fcs.Service, name string, p *PromptDeclaration, vars PromptVariables, executor ModelExecutor, w http.ResponseWriter) {
	user := ctx.Value(AuthenticatedUserKey).(string)
	if creditService != nil && p.Cost.Cost > 0 {
		creditGood, _, err := creditService.SubtractCredits(ctx, user)
		if err != nil {
			fmt.Printf("Failed to charge %d credits to user %s %v\n", p.Cost.Cost, user, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !creditGood {
			fmt.Printf("bad credit charge %d credits to user %s\n", p.Cost.Cost, user)
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
	}
	model_context, response, err := executor(p, vars)
	if err != nil {
		if creditService != nil && p.Cost.Cost > 0 {
			reterr := creditService.RefundCredits(ctx, user)
			if reterr != nil {
				fmt.Printf("Failed to return %d credits to user %s %v\n", p.Cost.Cost, user, reterr)
			}
		}
		fmt.Printf("Failed to process anthropic prompt %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	prompt_context := PromptContext{
		Prompt:       name,
		ModelContext: model_context,
	}

	contextJson, err := json.Marshal(prompt_context)
	if err != nil {
		fmt.Printf("failed to marshal context %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ret, err := MakeResult(contextJson, response)
	if err != nil {
		fmt.Printf("failed to make result %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	jsonResponse, err := json.Marshal(ret)
	if err != nil {
		fmt.Printf("failed to marshal response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(jsonResponse); err != nil {
		fmt.Printf("failed to write response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func setupcors() *cors.Cors {
	originsenv := os.Getenv("CORS_ORIGINS")
	if len(originsenv) > 0 {
		origins := strings.Split(originsenv, ",")
		options := cors.Options{
			AllowedOrigins:   origins,
			AllowedMethods:   []string{http.MethodPost},
			AllowCredentials: true,
			AllowedHeaders:   []string{"authorization"},
		}
		return cors.New(options)
	} else {
		return cors.Default()
	}
}

func continuanceConstructor(name string, p *PromptDeclaration) http.HandlerFunc {
	creditService := fcs.NewService(p.Cost, firebaseURL)
	return func(w http.ResponseWriter, r *http.Request) {
		if p.Service != Anthropic {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		vars := CollectContinuanceVariables(r)
		executor := AnthropicContinuePrompt
		runFunc(r.Context(), creditService, name, p, vars, executor, w)
	}
}

func continuance(w http.ResponseWriter, r *http.Request) {
	contextJson := r.FormValue("CONTEXT")
	var promptContext PromptContext
	if len(contextJson) <= 0 {
		fmt.Printf("CONTEXT not set\n")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := json.Unmarshal([]byte(contextJson), &promptContext)
	if err != nil {
		fmt.Printf("could not unpack CONTEXT\n")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	prompt, pok := prompts[promptContext.Prompt]
	if !pok {
		fmt.Printf("no prompt %s exists\n", promptContext.Prompt)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	NewTokenMiddleware(continuanceConstructor(promptContext.Prompt, &prompt), prompt.RequiredScope).ServeHTTP(w, r)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/continue", continuance)
	for k, v := range prompts {
		path := fmt.Sprintf("/v1/prompt/%s", k)
		mux.HandleFunc(path, NewTokenMiddleware(constructPromptHandler(k, &v), v.RequiredScope).ServeHTTP)
	}

	corsobj := setupcors()
	handler := corsobj.Handler(mux)

	if err := http.ListenAndServe("0.0.0.0:8080", handler); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

package main

import (
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

func constructPromptHandler(p *PromptDeclaration) http.HandlerFunc {
	creditService := fcs.NewService(p.Cost, firebaseURL)
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(AuthenticatedUserKey).(string)
		if p.Service != Anthropic {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		if p.Cost.Cost > 0 {
			creditGood, _, err := creditService.SubtractCredits(r.Context(), user)
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
		context, response, err := AnthropicProcessPrompt(r, p)
		if err != nil {
			if p.Cost.Cost > 0 {
				reterr := creditService.RefundCredits(r.Context(), user)
				if reterr != nil {
					fmt.Printf("Failed to return %d credits to user %s %v\n", p.Cost.Cost, user, reterr)
				}
			}
			fmt.Printf("Failed to process anthropic prompt %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ret, err := MakeResult(context, response)
		if err != nil {
			fmt.Printf("failed to make result %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
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

func main() {
	mux := http.NewServeMux()

	for k, v := range prompts {
		path := fmt.Sprintf("/v1/prompt/%s", k)
		mux.HandleFunc(path, NewTokenMiddleware(constructPromptHandler(&v), v.RequiredScope).ServeHTTP)
	}

	corsobj := setupcors()
	handler := corsobj.Handler(mux)

	if err := http.ListenAndServe("0.0.0.0:8080", handler); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

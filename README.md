# simple-prompt-service

## Environment Variables

The following environment variables are required to run the service:

| Variable | Description |
|----------|-------------|
| `PROMPTS` | JSON configuration containing prompt declarations for different endpoints. Required for service initialization. |
| `FIREBASE_DB_URL` | URL for the Firebase database connection. Required for service operation. |
| `CORS_ORIGINS` | Comma-separated list of allowed CORS origins. Optional - defaults to CORS default settings if not set. |
| `CONTEXT_KEY` | 32-character encryption key used for context encryption/decryption. Must be exactly 32 characters long. |
| `ANTHROPIC_API_KEY` | API key for Anthropic's Claude service. Required if using Anthropic prompts. |
| `TOKEN_VALIDATION_URL` | URL endpoint used to validate authentication tokens. Required for token validation. |
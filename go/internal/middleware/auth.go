package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// APIKeys stores valid API keys
var APIKeys map[string]bool

// WebhookSecret stores the webhook secret
var WebhookSecret string

// InitAuth initializes authentication settings
func InitAuth() {
	APIKeys = make(map[string]bool)
	
	// Load API keys from environment
	apiKeysEnv := os.Getenv("ACP_API_KEYS")
	if apiKeysEnv == "" {
		apiKeysEnv = "test_key_123,another_key"
	}
	
	keys := strings.Split(apiKeysEnv, ",")
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" {
			APIKeys[key] = true
		}
	}
	
	// Load webhook secret
	WebhookSecret = os.Getenv("ACP_WEBHOOK_SECRET")
	if WebhookSecret == "" {
		WebhookSecret = "whsec_123"
	}
}

// RequireAPIKey middleware that requires API key authentication
func RequireAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"missing bearer token"}`, http.StatusUnauthorized)
			return
		}
		
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if !APIKeys[token] {
			http.Error(w, `{"error":"invalid api key"}`, http.StatusUnauthorized)
			return
		}
		
		next.ServeHTTP(w, r)
	}
}

// IsSandbox checks if the request is in sandbox mode
func IsSandbox(r *http.Request) bool {
	sandboxHeader := r.Header.Get("X-ACP-Sandbox")
	if strings.ToLower(sandboxHeader) == "true" {
		return true
	}
	
	sandboxQuery := r.URL.Query().Get("sandbox")
	if sandboxQuery == "1" {
		return true
	}
	
	return false
}

// VerifyWebhookSignature verifies HMAC signature for webhooks
func VerifyWebhookSignature(signature string, body []byte) error {
	if signature == "" {
		return fmt.Errorf("missing signature")
	}
	
	// Parse signature: "t=<ts>,v1=<hex>"
	parts := strings.Split(signature, ",")
	var v1 string
	for _, part := range parts {
		if strings.HasPrefix(part, "v1=") {
			v1 = strings.TrimPrefix(part, "v1=")
			break
		}
	}
	
	if v1 == "" {
		return fmt.Errorf("bad signature header")
	}
	
	// Calculate HMAC
	mac := hmac.New(sha256.New, []byte(WebhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	
	// Compare signatures
	if !hmac.Equal([]byte(expectedMAC), []byte(v1)) {
		return fmt.Errorf("invalid signature")
	}
	
	return nil
}

// RequireWebhookSignature middleware that requires webhook signature verification
func RequireWebhookSignature(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		signature := r.Header.Get("X-Acp-Signature")
		
		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
			return
		}
		
		// Verify signature
		if err := VerifyWebhookSignature(signature, body); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusUnauthorized)
			return
		}
		
		// Reset body for next handler
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		next.ServeHTTP(w, r)
	}
}

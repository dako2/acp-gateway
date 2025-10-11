package main

import (
	"log"
	"net/http"
	"os"

	"acp-gateway-go/internal/handlers"
	"acp-gateway-go/internal/middleware"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize authentication
	middleware.InitAuth()
	
	// Create router
	r := mux.NewRouter()
	
	// Add CORS middleware
	corsHandler := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins([]string{"*"}),
		gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-ACP-Sandbox"}),
	)(r)
	
	// Add logging middleware
	loggedRouter := gorillaHandlers.LoggingHandler(os.Stdout, corsHandler)
	
	// Health check endpoint (no auth required)
	r.HandleFunc("/healthz", handlers.HealthHandler).Methods("GET")
	
	// ACP endpoints (require API key)
	acp := r.PathPrefix("/acp/v1").Subrouter()
	acp.HandleFunc("/products", middleware.RequireAPIKey(handlers.ProductsHandler)).Methods("GET")
	acp.HandleFunc("/intent", middleware.RequireAPIKey(handlers.IntentHandler)).Methods("POST")
	acp.HandleFunc("/checkout", middleware.RequireAPIKey(handlers.CheckoutHandler)).Methods("POST")
	
	// Merchant endpoints (require API key)
	merchant := r.PathPrefix("/merchant").Subrouter()
	merchant.HandleFunc("/register", middleware.RequireAPIKey(handlers.MerchantRegisterHandler)).Methods("POST")
	merchant.HandleFunc("/feed", middleware.RequireAPIKey(handlers.MerchantFeedHandler)).Methods("POST")
	
	// Webhook endpoints (require signature verification)
	r.HandleFunc("/webhooks/acp", middleware.RequireWebhookSignature(handlers.WebhooksHandler)).Methods("POST")
	
	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	
	log.Printf("ACP Gateway Go server starting on port %s", port)
	log.Printf("Environment variables:")
	log.Printf("  ACP_API_KEYS: %s", os.Getenv("ACP_API_KEYS"))
	log.Printf("  ACP_WEBHOOK_SECRET: %s", os.Getenv("ACP_WEBHOOK_SECRET"))
	log.Printf("  PORT: %s", port)
	
	// Start server
	if err := http.ListenAndServe(":"+port, loggedRouter); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

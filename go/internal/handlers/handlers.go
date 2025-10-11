package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"acp-gateway-go/internal/middleware"
	"acp-gateway-go/internal/models"
)

// Storage represents in-memory storage for the demo
type Storage struct {
	Products  []models.Product
	Carts     map[string]models.Cart
	Orders    map[string]models.Order
	Merchants map[string]models.Merchant
	AuditLogs []models.AuditLog
}

// NewStorage creates a new storage instance with sample data
func NewStorage() *Storage {
	storage := &Storage{
		Products:  make([]models.Product, 0),
		Carts:     make(map[string]models.Cart),
		Orders:    make(map[string]models.Order),
		Merchants: make(map[string]models.Merchant),
		AuditLogs: make([]models.AuditLog, 0),
	}
	
	// Add sample products
	storage.Products = []models.Product{
		{
			ID:          "sku_123",
			Title:       "Work Boots",
			Description: "Durable work boots for construction and outdoor work",
			Variants: []models.ProductVariant{
				{
					ID:    "var_9",
					Size:  "9",
					Price: models.Money{Value: 119.99, Currency: "USD"},
				},
			},
			Price:     models.Money{Value: 119.99, Currency: "USD"},
			Available: true,
			Images:    []string{"https://example.com/boots.jpg"},
			Tags:      []string{"work", "boots", "durable"},
		},
		{
			ID:          "sku_456",
			Title:       "Safety Helmet",
			Description: "ANSI approved safety helmet",
			Variants: []models.ProductVariant{
				{
					ID:    "var_l",
					Size:  "Large",
					Price: models.Money{Value: 45.99, Currency: "USD"},
				},
			},
			Price:     models.Money{Value: 45.99, Currency: "USD"},
			Available: true,
			Images:    []string{"https://example.com/helmet.jpg"},
			Tags:      []string{"safety", "helmet", "ansi"},
		},
	}
	
	return storage
}

// Audit logs an audit event
func (s *Storage) Audit(event string, data map[string]interface{}) {
	s.AuditLogs = append(s.AuditLogs, models.AuditLog{
		Timestamp: time.Now(),
		Event:     event,
		Data:      data,
	})
}

// Global storage instance
var storage = NewStorage()

// HealthHandler handles health check requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.HealthResponse{OK: true})
}

// ProductsHandler handles product listing and search
func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 20
	offset := 0
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}
	
	// Filter products
	var filteredProducts []models.Product
	if query != "" {
		queryLower := strings.ToLower(query)
		for _, product := range storage.Products {
			if strings.Contains(strings.ToLower(product.Title), queryLower) ||
				strings.Contains(strings.ToLower(product.ID), queryLower) ||
				strings.Contains(strings.ToLower(product.Description), queryLower) {
				filteredProducts = append(filteredProducts, product)
			}
		}
	} else {
		filteredProducts = storage.Products
	}
	
	// Apply pagination
	total := len(filteredProducts)
	end := offset + limit
	if end > total {
		end = total
	}
	
	var items []models.Product
	if offset < total {
		items = filteredProducts[offset:end]
	}
	
	response := models.APIResponse{
		OK:    true,
		Items: items,
		Total: total,
	}
	
	json.NewEncoder(w).Encode(response)
}

// IntentHandler handles ACP intent requests
func IntentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	sandbox := middleware.IsSandbox(r)
	
	var intent models.Intent
	if err := json.NewDecoder(r.Body).Decode(&intent); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	
	// Initialize or get cart
	cartID := intent.Payload.Context.SessionID
	if cartID == "" {
		cartID = fmt.Sprintf("cart_%d", time.Now().UnixNano())
	}
	
	cart, exists := storage.Carts[cartID]
	if !exists {
		cart = models.Cart{
			ID:    cartID,
			Items: make([]models.CartItem, 0),
		}
	}
	
	// Process intent action
	switch intent.Payload.Action {
	case "add_to_cart":
		for _, item := range intent.Payload.Items {
			// Find product
			var product *models.Product
			for _, p := range storage.Products {
				if p.ID == item.ID {
					product = &p
					break
				}
			}
			
			if product == nil {
				http.Error(w, fmt.Sprintf(`{"error":"product %s not found"}`, item.ID), http.StatusNotFound)
				return
			}
			
			// Determine price
			price := product.Price
			if item.Price != nil {
				price = *item.Price
			}
			
			cartItem := models.CartItem{
				ID:        item.ID,
				VariantID: item.VariantID,
				Quantity:  item.Quantity,
				Price:     price,
			}
			
			cart.Items = append(cart.Items, cartItem)
		}
		
	case "select_variant":
		for _, item := range intent.Payload.Items {
			for i := range cart.Items {
				if cart.Items[i].ID == item.ID {
					cart.Items[i].VariantID = item.VariantID
					break
				}
			}
		}
		
	case "set_quantity":
		for _, item := range intent.Payload.Items {
			for i := range cart.Items {
				if cart.Items[i].ID == item.ID {
					cart.Items[i].Quantity = item.Quantity
					break
				}
			}
		}
	}
	
	// Calculate subtotal
	var subtotal float64
	for _, item := range cart.Items {
		subtotal += item.Price.Value * float64(item.Quantity)
	}
	
	cart.Subtotal = models.Money{
		Value:    math.Round(subtotal*100) / 100,
		Currency: "USD",
	}
	
	storage.Carts[cartID] = cart
	
	result := models.IntentResult{
		Type:    "acp.intent.result",
		OK:      true,
		Sandbox: sandbox,
		Cart:    &cart,
		Next: &models.NextAction{
			Action:   "checkout",
			Endpoint: "/acp/v1/checkout",
		},
	}
	
	// Audit the intent
	auditData := map[string]interface{}{
		"sandbox": sandbox,
		"intent":  intent,
		"result":  result,
	}
	storage.Audit("intent", auditData)
	
	json.NewEncoder(w).Encode(result)
}

// CheckoutHandler handles checkout requests
func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	sandbox := middleware.IsSandbox(r)
	
	var req models.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	
	cart, exists := storage.Carts[req.CartID]
	if !exists {
		http.Error(w, `{"error":"cart not found"}`, http.StatusNotFound)
		return
	}
	
	orderID := fmt.Sprintf("ord_%d", time.Now().UnixNano())
	
	order := models.Order{
		ID:              orderID,
		CartID:          req.CartID,
		Email:           "",
		Items:           cart.Items,
		Sandbox:         sandbox,
		Status:          "created",
		ShippingAddress: req.ShippingAddress,
		Payment:         req.Payment,
		CreatedAt:       time.Now(),
	}
	
	if req.Buyer != nil {
		order.Email = req.Buyer.Email
	}
	
	storage.Orders[orderID] = order
	
	paymentStatus := "requires_action"
	if !sandbox {
		paymentStatus = "succeeded"
	}
	
	response := models.CheckoutResponse{
		OK:                  true,
		Sandbox:             sandbox,
		OrderID:             orderID,
		PaymentIntentStatus: paymentStatus,
		RedirectURL:         fmt.Sprintf("https://merchant.example/checkout/%s", orderID),
	}
	
	// Audit the checkout
	auditData := map[string]interface{}{
		"sandbox": sandbox,
		"req":     req,
		"resp":    response,
	}
	storage.Audit("checkout", auditData)
	
	json.NewEncoder(w).Encode(response)
}

// MerchantRegisterHandler handles merchant registration
func MerchantRegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var merchant models.Merchant
	if err := json.NewDecoder(r.Body).Decode(&merchant); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	
	merchantID := merchant.Domain
	merchant.ID = merchantID
	
	storage.Merchants[merchantID] = merchant
	
	// Audit the registration
	auditData := map[string]interface{}{
		"merchant": merchant,
	}
	storage.Audit("merchant.register", auditData)
	
	response := map[string]interface{}{
		"ok":          true,
		"merchant_id": merchantID,
	}
	
	json.NewEncoder(w).Encode(response)
}

// MerchantFeedHandler handles merchant feed uploads
func MerchantFeedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	url := r.URL.Query().Get("url")
	
	// Handle file upload
	if r.MultipartForm != nil {
		file, _, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			content, _ := io.ReadAll(file)
			if len(content) > 5000 {
				content = content[:5000] // Truncate for demo
			}
			
			// Audit the upload
			auditData := map[string]interface{}{
				"size": len(content),
			}
			storage.Audit("merchant.feed.upload", auditData)
		}
	}
	
	if url != "" {
		// Audit the URL
		auditData := map[string]interface{}{
			"url": url,
		}
		storage.Audit("merchant.feed.url", auditData)
	}
	
	// Demo: keep only first product
	if len(storage.Products) > 1 {
		storage.Products = storage.Products[:1]
	}
	
	response := map[string]interface{}{
		"ok":        true,
		"ingested":  len(storage.Products),
	}
	
	json.NewEncoder(w).Encode(response)
}

// WebhooksHandler handles webhook requests
func WebhooksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	
	// Audit the webhook
	storage.Audit("webhook", payload)
	
	response := map[string]interface{}{
		"ok": true,
	}
	
	json.NewEncoder(w).Encode(response)
}

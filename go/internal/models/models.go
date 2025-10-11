package models

import (
	"time"
)

// Money represents a monetary amount
type Money struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

// Address represents a shipping or billing address
type Address struct {
	Name       string `json:"name"`
	LineOne    string `json:"line_one"`
	LineTwo    string `json:"line_two,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
}

// Buyer represents buyer information
type Buyer struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// Item represents a product item
type Item struct {
	ID         string  `json:"id"`
	VariantID  *string `json:"variant_id,omitempty"`  // ACP extension
	Quantity   int     `json:"quantity"`
	Price      *Money  `json:"price,omitempty"`       // ACP extension
}

// CartItem represents an item in a cart
type CartItem struct {
	ID         string  `json:"id"`
	VariantID  *string `json:"variant_id,omitempty"`
	Quantity   int     `json:"quantity"`
	Price      Money   `json:"price"`
}

// Cart represents a shopping cart
type Cart struct {
	ID       string     `json:"id"`
	Items    []CartItem `json:"items"`
	Subtotal Money      `json:"subtotal"`
	Total    *Money     `json:"total,omitempty"`
	Taxes    *Money     `json:"taxes,omitempty"`
	Shipping *Money     `json:"shipping,omitempty"`
}

// Product represents a product in the catalog
type Product struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description,omitempty"`
	Variants    []ProductVariant `json:"variants"`
	Price       Money            `json:"price"`
	Available   bool             `json:"available"`
	Images      []string         `json:"images,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
}

// ProductVariant represents a product variant
type ProductVariant struct {
	ID    string `json:"id"`
	Size  string `json:"size,omitempty"`
	Price Money  `json:"price"`
}

// Payment represents payment information
type Payment struct {
	Method string                 `json:"method"`
	Token  string                 `json:"token,omitempty"`
	Card   map[string]interface{} `json:"card,omitempty"`
}

// IntentContext represents context for an intent
type IntentContext struct {
	SessionID string            `json:"session_id,omitempty"`
	Locale    string            `json:"locale,omitempty"`
	UserID    string            `json:"user_id,omitempty"`
	Extra     map[string]string `json:"-"` // For additional context
}

// IntentPayload represents the payload of an intent
type IntentPayload struct {
	Action string        `json:"action"`
	Items  []Item        `json:"items"`
	Notes  string        `json:"notes,omitempty"`
	Context IntentContext `json:"context,omitempty"`
}

// Intent represents an ACP intent
type Intent struct {
	Type    string        `json:"type"`
	Actor   string        `json:"actor"`
	Payload IntentPayload `json:"payload"`
}

// NextAction represents the next action to take
type NextAction struct {
	Action   string                 `json:"action"`
	Endpoint string                 `json:"endpoint"`
	Method   string                 `json:"method,omitempty"`
	Params   map[string]interface{} `json:"params,omitempty"`
}

// IntentResult represents the result of processing an intent
type IntentResult struct {
	Type    string     `json:"type"`
	OK      bool       `json:"ok"`
	Sandbox bool       `json:"sandbox"`
	Cart    *Cart      `json:"cart,omitempty"`
	Next    *NextAction `json:"next,omitempty"`
	Error   string     `json:"error,omitempty"`
}

// CheckoutRequest represents a checkout request
type CheckoutRequest struct {
	CartID         string    `json:"cart_id"`
	Buyer          *Buyer    `json:"buyer,omitempty"`
	ShippingAddress *Address `json:"shipping_address,omitempty"`
	BillingAddress *Address `json:"billing_address,omitempty"`
	Payment        *Payment  `json:"payment,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CheckoutResponse represents a checkout response
type CheckoutResponse struct {
	OK                   bool    `json:"ok"`
	Sandbox              bool    `json:"sandbox"`
	OrderID              string  `json:"order_id,omitempty"`
	CheckoutID           string  `json:"checkout_id,omitempty"`
	PaymentIntentStatus  string  `json:"payment_intent_status,omitempty"`
	RedirectURL          string  `json:"redirect_url,omitempty"`
	ClientSecret         string  `json:"client_secret,omitempty"`
	Total                *Money  `json:"total,omitempty"`
	Error                string  `json:"error,omitempty"`
}

// Merchant represents a merchant
type Merchant struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Domain   string                 `json:"domain"`
	APIBase  string                 `json:"api_base,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Order represents an order
type Order struct {
	ID              string    `json:"id"`
	CartID          string    `json:"cart_id"`
	Email           string    `json:"email,omitempty"`
	Items           []CartItem `json:"items"`
	Sandbox         bool      `json:"sandbox"`
	Status          string    `json:"status"`
	ShippingAddress *Address  `json:"shipping_address,omitempty"`
	Payment         *Payment  `json:"payment,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	Timestamp time.Time              `json:"ts"`
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	OK    bool        `json:"ok"`
	Items interface{} `json:"items,omitempty"`
	Total int         `json:"total,omitempty"`
	Error string      `json:"error,omitempty"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	OK bool `json:"ok"`
}

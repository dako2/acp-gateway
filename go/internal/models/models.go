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

// Product represents a product in the catalog (legacy model - use ProductFeed for full OpenAI spec compliance)
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

// ProductVariant represents a product variant (legacy model - use ProductFeed for full OpenAI spec compliance)
type ProductVariant struct {
	ID    string `json:"id"`
	Size  string `json:"size,omitempty"`
	Price Money  `json:"price"`
}

// EnhancedProduct represents a product with OpenAI Product Feed Specification compliance
type EnhancedProduct struct {
	ProductFeed ProductFeed `json:"product_feed"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	MerchantID  string      `json:"merchant_id"`
	Status      string      `json:"status"` // active, inactive, pending_review
}

// ProductSearchRequest represents a request to search products
type ProductSearchRequest struct {
	Query       string                 `json:"query,omitempty"`
	Category    string                 `json:"category,omitempty"`
	Brand       string                 `json:"brand,omitempty"`
	MinPrice    *Money                 `json:"min_price,omitempty"`
	MaxPrice    *Money                 `json:"max_price,omitempty"`
	Availability string                `json:"availability,omitempty"`
	MerchantID  string                 `json:"merchant_id,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Offset      int                    `json:"offset,omitempty"`
	SortBy      string                 `json:"sort_by,omitempty"`
	SortOrder   string                 `json:"sort_order,omitempty"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
}

// ProductSearchResponse represents the response from product search
type ProductSearchResponse struct {
	Products []EnhancedProduct `json:"products"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	HasMore  bool              `json:"has_more"`
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

// OpenAI Flags - Controls product visibility in ChatGPT
type OpenAIFlags struct {
	EnableSearch  bool `json:"enable_search" validate:"required"`
	EnableCheckout bool `json:"enable_checkout" validate:"required"`
}

// Basic Product Data
type BasicProductData struct {
	ID          string `json:"id" validate:"required,max=100"`
	GTIN        string `json:"gtin,omitempty" validate:"omitempty,len=8|len=12|len=13|len=14"`
	MPN         string `json:"mpn,omitempty" validate:"max=70"`
	Title       string `json:"title" validate:"required,max=150"`
	Description string `json:"description" validate:"required,max=5000"`
	Link        string `json:"link" validate:"required,url"`
}

// Item Information
type ItemInformation struct {
	Condition   string `json:"condition,omitempty" validate:"omitempty,oneof=new refurbished used"`
	Category    string `json:"product_category" validate:"required"`
	Brand       string `json:"brand,omitempty" validate:"max=70"`
	Material    string `json:"material,omitempty" validate:"max=100"`
	Dimensions  string `json:"dimensions,omitempty"`
	Length      string `json:"length,omitempty"`
	Width       string `json:"width,omitempty"`
	Height      string `json:"height,omitempty"`
	Weight      string `json:"weight,omitempty"`
	AgeGroup    string `json:"age_group,omitempty" validate:"omitempty,oneof=newborn infant toddler kids adult"`
}

// Media
type Media struct {
	ImageLink           string   `json:"image_link" validate:"required,url"`
	AdditionalImageLink []string `json:"additional_image_link,omitempty"`
	VideoLink           string   `json:"video_link,omitempty" validate:"omitempty,url"`
	Model3DLink         string   `json:"model_3d_link,omitempty" validate:"omitempty,url"`
}

// Price and Promotions
type PricePromotions struct {
	Price                    Money   `json:"price" validate:"required"`
	ApplicableTaxesFees      *Money  `json:"applicable_taxes_fees,omitempty"`
	SalePrice                *Money  `json:"sale_price,omitempty"`
	SalePriceEffectiveDate   string  `json:"sale_price_effective_date,omitempty"`
	UnitPricingMeasure       string  `json:"unit_pricing_measure,omitempty"`
	BaseMeasure              string  `json:"base_measure,omitempty"`
	PricingTrend             string  `json:"pricing_trend,omitempty" validate:"max=80"`
}

// Availability and Inventory
type AvailabilityInventory struct {
	Availability      string `json:"availability" validate:"required,oneof=in_stock out_of_stock preorder"`
	AvailabilityDate  string `json:"availability_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	InventoryQuantity int    `json:"inventory_quantity" validate:"min=0"`
	ExpirationDate    string `json:"expiration_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	PickupMethod      string `json:"pickup_method,omitempty" validate:"omitempty,oneof=in_store reserve not_supported"`
	PickupSLA         string `json:"pickup_sla,omitempty"`
}

// Variants
type Variants struct {
	ItemGroupID            string `json:"item_group_id,omitempty" validate:"max=70"`
	ItemGroupTitle         string `json:"item_group_title,omitempty" validate:"max=150"`
	Color                  string `json:"color,omitempty" validate:"max=40"`
	Size                   string `json:"size,omitempty" validate:"max=20"`
	SizeSystem             string `json:"size_system,omitempty" validate:"omitempty,len=2"`
	Gender                 string `json:"gender,omitempty" validate:"omitempty,oneof=male female unisex"`
	OfferID                string `json:"offer_id,omitempty"`
	CustomVariant1Category string `json:"custom_variant1_category,omitempty"`
	CustomVariant1Option   string `json:"custom_variant1_option,omitempty"`
	CustomVariant2Category string `json:"custom_variant2_category,omitempty"`
	CustomVariant2Option   string `json:"custom_variant2_option,omitempty"`
	CustomVariant3Category string `json:"custom_variant3_category,omitempty"`
	CustomVariant3Option   string `json:"custom_variant3_option,omitempty"`
}

// Fulfillment
type Fulfillment struct {
	Shipping         []string `json:"shipping,omitempty"`
	DeliveryEstimate string   `json:"delivery_estimate,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

// Merchant Information
type MerchantInfo struct {
	SellerName           string `json:"seller_name" validate:"required,max=70"`
	SellerURL            string `json:"seller_url" validate:"required,url"`
	SellerPrivacyPolicy  string `json:"seller_privacy_policy,omitempty" validate:"omitempty,url"`
	SellerTOS            string `json:"seller_tos,omitempty" validate:"omitempty,url"`
}

// Returns
type Returns struct {
	ReturnPolicy string `json:"return_policy" validate:"required,url"`
	ReturnWindow int    `json:"return_window" validate:"min=1"`
}

// Performance Signals
type PerformanceSignals struct {
	PopularityScore float64 `json:"popularity_score,omitempty" validate:"omitempty,min=0,max=5"`
	ReturnRate      float64 `json:"return_rate,omitempty" validate:"omitempty,min=0,max=100"`
}

// Compliance
type Compliance struct {
	Warning        string `json:"warning,omitempty"`
	WarningURL     string `json:"warning_url,omitempty" validate:"omitempty,url"`
	AgeRestriction int    `json:"age_restriction,omitempty" validate:"omitempty,min=1"`
}

// Reviews and Q&A
type ReviewsQA struct {
	ProductReviewCount int     `json:"product_review_count,omitempty" validate:"omitempty,min=0"`
	ProductReviewRating float64 `json:"product_review_rating,omitempty" validate:"omitempty,min=0,max=5"`
	StoreReviewCount   int     `json:"store_review_count,omitempty" validate:"omitempty,min=0"`
	StoreReviewRating  float64 `json:"store_review_rating,omitempty" validate:"omitempty,min=0,max=5"`
	QAndA              string  `json:"q_and_a,omitempty"`
	RawReviewData      string  `json:"raw_review_data,omitempty"`
}

// Related Products
type RelatedProducts struct {
	RelatedProductIDs []string `json:"related_product_id,omitempty"`
	RelationshipType  string   `json:"relationship_type,omitempty" validate:"omitempty,oneof=part_of_set required_part often_bought_with substitute different_brand accessory"`
}

// Geo Tagging
type GeoTagging struct {
	GeoPrice        *Money `json:"geo_price,omitempty"`
	GeoAvailability string `json:"geo_availability,omitempty"`
}

// ProductFeed represents a complete product feed entry according to OpenAI Product Feed Specification
type ProductFeed struct {
	OpenAIFlags            OpenAIFlags            `json:"openai_flags"`
	BasicData              BasicProductData       `json:"basic_data"`
	ItemInfo               ItemInformation        `json:"item_info"`
	Media                  Media                  `json:"media"`
	PricePromotions        PricePromotions        `json:"price_promotions"`
	AvailabilityInventory  AvailabilityInventory  `json:"availability_inventory"`
	Variants               Variants               `json:"variants"`
	Fulfillment            Fulfillment            `json:"fulfillment"`
	MerchantInfo           MerchantInfo           `json:"merchant_info"`
	Returns                Returns                `json:"returns"`
	PerformanceSignals     PerformanceSignals     `json:"performance_signals,omitempty"`
	Compliance             Compliance             `json:"compliance,omitempty"`
	ReviewsQA              ReviewsQA              `json:"reviews_qa,omitempty"`
	RelatedProducts        RelatedProducts        `json:"related_products,omitempty"`
	GeoTagging             GeoTagging             `json:"geo_tagging,omitempty"`
}

// FeedIngestionRequest represents a request to ingest a product feed
type FeedIngestionRequest struct {
	MerchantID string      `json:"merchant_id" validate:"required"`
	FeedData   []ProductFeed `json:"feed_data" validate:"required"`
	Format     string      `json:"format" validate:"required,oneof=json csv tsv xml"`
	Version    string      `json:"version,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// FeedIngestionResponse represents the response from feed ingestion
type FeedIngestionResponse struct {
	OK           bool   `json:"ok"`
	ProcessedCount int   `json:"processed_count"`
	ErrorCount   int    `json:"error_count"`
	Errors       []string `json:"errors,omitempty"`
	FeedID       string `json:"feed_id,omitempty"`
	Message      string `json:"message,omitempty"`
}

// FeedValidationResult represents validation results for a product feed
type FeedValidationResult struct {
	Valid         bool                   `json:"valid"`
	Errors        []ValidationError      `json:"errors,omitempty"`
	Warnings      []ValidationWarning    `json:"warnings,omitempty"`
	ProductCount  int                    `json:"product_count"`
	ValidProducts int                    `json:"valid_products"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field       string `json:"field"`
	Message     string `json:"message"`
	ProductID   string `json:"product_id,omitempty"`
	Severity    string `json:"severity"`
	Code        string `json:"code"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field     string `json:"field"`
	Message   string `json:"message"`
	ProductID string `json:"product_id,omitempty"`
	Code      string `json:"code"`
}

// FeedStatus represents the status of a product feed
type FeedStatus struct {
	FeedID       string    `json:"feed_id"`
	MerchantID   string    `json:"merchant_id"`
	Status       string    `json:"status"` // pending, processing, completed, failed
	ProcessedAt  time.Time `json:"processed_at"`
	ProductCount int       `json:"product_count"`
	ErrorCount   int       `json:"error_count"`
	LastUpdated  time.Time `json:"last_updated"`
}

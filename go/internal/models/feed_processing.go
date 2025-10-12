package models

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"
)

// FeedProcessor handles parsing and processing of different feed formats
type FeedProcessor struct {
	validator *Validator
}

// NewFeedProcessor creates a new feed processor
func NewFeedProcessor() *FeedProcessor {
	return &FeedProcessor{
		validator: NewValidator(),
	}
}

// ProcessFeed processes a product feed in the specified format
func (fp *FeedProcessor) ProcessFeed(format string, data []byte) (*FeedIngestionRequest, error) {
	switch strings.ToLower(format) {
	case "json":
		return fp.processJSONFeed(data)
	case "csv":
		return fp.processCSVFeed(data)
	case "tsv":
		return fp.processTSVFeed(data)
	case "xml":
		return fp.processXMLFeed(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// processJSONFeed processes a JSON formatted feed
func (fp *FeedProcessor) processJSONFeed(data []byte) (*FeedIngestionRequest, error) {
	var request FeedIngestionRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, fmt.Errorf("failed to parse JSON feed: %w", err)
	}
	return &request, nil
}

// processCSVFeed processes a CSV formatted feed
func (fp *FeedProcessor) processCSVFeed(data []byte) (*FeedIngestionRequest, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV feed: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV feed must have at least a header row and one data row")
	}

	header := records[0]
	products := make([]ProductFeed, 0, len(records)-1)

	for i, record := range records[1:] {
		product, err := fp.csvRecordToProductFeed(header, record)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CSV record %d: %w", i+1, err)
		}
		products = append(products, *product)
	}

	// For CSV feeds, we'll need to extract merchant ID from metadata or use a default
	merchantID := "csv_merchant" // This should be provided in the request context

	return &FeedIngestionRequest{
		MerchantID: merchantID,
		FeedData:   products,
		Format:     "csv",
		Version:    "1.0",
		Metadata: map[string]interface{}{
			"source": "csv_upload",
			"processed_at": time.Now().UTC(),
		},
	}, nil
}

// processTSVFeed processes a TSV (Tab-Separated Values) formatted feed
func (fp *FeedProcessor) processTSVFeed(data []byte) (*FeedIngestionRequest, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = '\t' // Use tab as delimiter
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse TSV feed: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("TSV feed must have at least a header row and one data row")
	}

	header := records[0]
	products := make([]ProductFeed, 0, len(records)-1)

	for i, record := range records[1:] {
		product, err := fp.csvRecordToProductFeed(header, record)
		if err != nil {
			return nil, fmt.Errorf("failed to parse TSV record %d: %w", i+1, err)
		}
		products = append(products, *product)
	}

	merchantID := "tsv_merchant" // This should be provided in the request context

	return &FeedIngestionRequest{
		MerchantID: merchantID,
		FeedData:   products,
		Format:     "tsv",
		Version:    "1.0",
		Metadata: map[string]interface{}{
			"source": "tsv_upload",
			"processed_at": time.Now().UTC(),
		},
	}, nil
}

// processXMLFeed processes an XML formatted feed
func (fp *FeedProcessor) processXMLFeed(data []byte) (*FeedIngestionRequest, error) {
	var xmlFeed struct {
		XMLName   xml.Name      `xml:"feed"`
		MerchantID string       `xml:"merchant_id,attr"`
		Products   []ProductFeed `xml:"product"`
	}

	if err := xml.Unmarshal(data, &xmlFeed); err != nil {
		return nil, fmt.Errorf("failed to parse XML feed: %w", err)
	}

	merchantID := xmlFeed.MerchantID
	if merchantID == "" {
		merchantID = "xml_merchant" // Default fallback
	}

	return &FeedIngestionRequest{
		MerchantID: merchantID,
		FeedData:   xmlFeed.Products,
		Format:     "xml",
		Version:    "1.0",
		Metadata: map[string]interface{}{
			"source": "xml_upload",
			"processed_at": time.Now().UTC(),
		},
	}, nil
}

// csvRecordToProductFeed converts a CSV record to a ProductFeed
func (fp *FeedProcessor) csvRecordToProductFeed(header, record []string) (*ProductFeed, error) {
	if len(header) != len(record) {
		return nil, fmt.Errorf("header and record length mismatch")
	}

	// Create a map for easy field access
	fieldMap := make(map[string]string)
	for i, field := range header {
		fieldMap[strings.ToLower(strings.TrimSpace(field))] = strings.TrimSpace(record[i])
	}

	product := &ProductFeed{
		OpenAIFlags: OpenAIFlags{
			EnableSearch:  fp.parseBool(fieldMap["enable_search"], true),
			EnableCheckout: fp.parseBool(fieldMap["enable_checkout"], false),
		},
		BasicData: BasicProductData{
			ID:          fieldMap["id"],
			GTIN:        fieldMap["gtin"],
			MPN:         fieldMap["mpn"],
			Title:       fieldMap["title"],
			Description: fieldMap["description"],
			Link:        fieldMap["link"],
		},
		ItemInfo: ItemInformation{
			Condition:  fieldMap["condition"],
			Category:   fieldMap["product_category"],
			Brand:      fieldMap["brand"],
			Material:   fieldMap["material"],
			Dimensions: fieldMap["dimensions"],
			Length:     fieldMap["length"],
			Width:      fieldMap["width"],
			Height:     fieldMap["height"],
			Weight:     fieldMap["weight"],
			AgeGroup:   fieldMap["age_group"],
		},
		Media: Media{
			ImageLink:           fieldMap["image_link"],
			AdditionalImageLink: fp.parseStringArray(fieldMap["additional_image_link"]),
			VideoLink:           fieldMap["video_link"],
			Model3DLink:         fieldMap["model_3d_link"],
		},
		PricePromotions: PricePromotions{
			Price: Money{
				Value:    fp.parseFloat(fieldMap["price_value"], 0),
				Currency: fieldMap["price_currency"],
			},
			SalePrice:                fp.parseOptionalMoney(fieldMap["sale_price_value"], fieldMap["sale_price_currency"]),
			SalePriceEffectiveDate:   fieldMap["sale_price_effective_date"],
			UnitPricingMeasure:       fieldMap["unit_pricing_measure"],
			BaseMeasure:              fieldMap["base_measure"],
			PricingTrend:             fieldMap["pricing_trend"],
		},
		AvailabilityInventory: AvailabilityInventory{
			Availability:      fieldMap["availability"],
			AvailabilityDate:  fieldMap["availability_date"],
			InventoryQuantity: fp.parseInt(fieldMap["inventory_quantity"], 0),
			ExpirationDate:    fieldMap["expiration_date"],
			PickupMethod:      fieldMap["pickup_method"],
			PickupSLA:         fieldMap["pickup_sla"],
		},
		Variants: Variants{
			ItemGroupID:            fieldMap["item_group_id"],
			ItemGroupTitle:         fieldMap["item_group_title"],
			Color:                  fieldMap["color"],
			Size:                   fieldMap["size"],
			SizeSystem:             fieldMap["size_system"],
			Gender:                 fieldMap["gender"],
			OfferID:                fieldMap["offer_id"],
			CustomVariant1Category: fieldMap["custom_variant1_category"],
			CustomVariant1Option:   fieldMap["custom_variant1_option"],
			CustomVariant2Category: fieldMap["custom_variant2_category"],
			CustomVariant2Option:   fieldMap["custom_variant2_option"],
			CustomVariant3Category: fieldMap["custom_variant3_category"],
			CustomVariant3Option:   fieldMap["custom_variant3_option"],
		},
		Fulfillment: Fulfillment{
			Shipping:         fp.parseStringArray(fieldMap["shipping"]),
			DeliveryEstimate: fieldMap["delivery_estimate"],
		},
		MerchantInfo: MerchantInfo{
			SellerName:          fieldMap["seller_name"],
			SellerURL:           fieldMap["seller_url"],
			SellerPrivacyPolicy: fieldMap["seller_privacy_policy"],
			SellerTOS:           fieldMap["seller_tos"],
		},
		Returns: Returns{
			ReturnPolicy: fieldMap["return_policy"],
			ReturnWindow: fp.parseInt(fieldMap["return_window"], 30),
		},
		PerformanceSignals: PerformanceSignals{
			PopularityScore: fp.parseFloat(fieldMap["popularity_score"], 0),
			ReturnRate:      fp.parseFloat(fieldMap["return_rate"], 0),
		},
		Compliance: Compliance{
			Warning:        fieldMap["warning"],
			WarningURL:     fieldMap["warning_url"],
			AgeRestriction: fp.parseInt(fieldMap["age_restriction"], 0),
		},
		ReviewsQA: ReviewsQA{
			ProductReviewCount: fp.parseInt(fieldMap["product_review_count"], 0),
			ProductReviewRating: fp.parseFloat(fieldMap["product_review_rating"], 0),
			StoreReviewCount:   fp.parseInt(fieldMap["store_review_count"], 0),
			StoreReviewRating:  fp.parseFloat(fieldMap["store_review_rating"], 0),
			QAndA:              fieldMap["q_and_a"],
			RawReviewData:      fieldMap["raw_review_data"],
		},
		RelatedProducts: RelatedProducts{
			RelatedProductIDs: fp.parseStringArray(fieldMap["related_product_id"]),
			RelationshipType:  fieldMap["relationship_type"],
		},
		GeoTagging: GeoTagging{
			GeoPrice:        fp.parseOptionalMoney(fieldMap["geo_price_value"], fieldMap["geo_price_currency"]),
			GeoAvailability: fieldMap["geo_availability"],
		},
	}

	return product, nil
}

// Helper functions for parsing CSV data

func (fp *FeedProcessor) parseBool(value string, defaultValue bool) bool {
	if value == "" {
		return defaultValue
	}
	switch strings.ToLower(value) {
	case "true", "1", "yes", "y":
		return true
	case "false", "0", "no", "n":
		return false
	default:
		return defaultValue
	}
}

func (fp *FeedProcessor) parseInt(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	// Simple integer parsing - in production, use strconv.Atoi with proper error handling
	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return defaultValue
	}
	return result
}

func (fp *FeedProcessor) parseFloat(value string, defaultValue float64) float64 {
	if value == "" {
		return defaultValue
	}
	var result float64
	if _, err := fmt.Sscanf(value, "%f", &result); err != nil {
		return defaultValue
	}
	return result
}

func (fp *FeedProcessor) parseStringArray(value string) []string {
	if value == "" {
		return nil
	}
	// Split by comma and clean up whitespace
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (fp *FeedProcessor) parseOptionalMoney(valueStr, currencyStr string) *Money {
	if valueStr == "" || currencyStr == "" {
		return nil
	}
	value := fp.parseFloat(valueStr, 0)
	if value == 0 {
		return nil
	}
	return &Money{
		Value:    value,
		Currency: currencyStr,
	}
}

// FeedStorage represents operations for storing and retrieving product feeds
type FeedStorage struct {
	feeds map[string]*FeedStatus
}

// NewFeedStorage creates a new feed storage instance
func NewFeedStorage() *FeedStorage {
	return &FeedStorage{
		feeds: make(map[string]*FeedStatus),
	}
}

// StoreFeed stores a feed status
func (fs *FeedStorage) StoreFeed(feedID string, status *FeedStatus) {
	fs.feeds[feedID] = status
}

// GetFeed retrieves a feed status
func (fs *FeedStorage) GetFeed(feedID string) (*FeedStatus, bool) {
	status, exists := fs.feeds[feedID]
	return status, exists
}

// ListFeeds returns all feed statuses
func (fs *FeedStorage) ListFeeds() []*FeedStatus {
	statuses := make([]*FeedStatus, 0, len(fs.feeds))
	for _, status := range fs.feeds {
		statuses = append(statuses, status)
	}
	return statuses
}

// UpdateFeedStatus updates the status of a feed
func (fs *FeedStorage) UpdateFeedStatus(feedID string, status string, productCount, errorCount int) error {
	if feed, exists := fs.feeds[feedID]; exists {
		feed.Status = status
		feed.ProductCount = productCount
		feed.ErrorCount = errorCount
		feed.LastUpdated = time.Now().UTC()
		if status == "completed" || status == "failed" {
			feed.ProcessedAt = time.Now().UTC()
		}
		return nil
	}
	return fmt.Errorf("feed not found: %s", feedID)
}

// FeedTransformer provides methods to transform between different product representations
type FeedTransformer struct{}

// NewFeedTransformer creates a new feed transformer
func NewFeedTransformer() *FeedTransformer {
	return &FeedTransformer{}
}

// ProductFeedToEnhancedProduct converts a ProductFeed to an EnhancedProduct
func (ft *FeedTransformer) ProductFeedToEnhancedProduct(feed ProductFeed, merchantID string) *EnhancedProduct {
	return &EnhancedProduct{
		ProductFeed: feed,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		MerchantID:  merchantID,
		Status:      "active",
	}
}

// EnhancedProductToProductFeed converts an EnhancedProduct back to a ProductFeed
func (ft *FeedTransformer) EnhancedProductToProductFeed(product *EnhancedProduct) ProductFeed {
	feed := product.ProductFeed
	// Update any fields that might have changed
	return feed
}

// ProductFeedToLegacyProduct converts a ProductFeed to the legacy Product model for backward compatibility
func (ft *FeedTransformer) ProductFeedToLegacyProduct(feed ProductFeed) *Product {
	// Extract the main image
	images := []string{feed.Media.ImageLink}
	images = append(images, feed.Media.AdditionalImageLink...)

	// Create variants from the feed data
	variants := []ProductVariant{}
	if feed.Variants.ItemGroupID != "" {
		// If this is a variant product, create a variant entry
		variant := ProductVariant{
			ID:    feed.BasicData.ID,
			Size:  feed.Variants.Size,
			Price: feed.PricePromotions.Price,
		}
		variants = append(variants, variant)
	}

	return &Product{
		ID:          feed.BasicData.ID,
		Title:       feed.BasicData.Title,
		Description: feed.BasicData.Description,
		Variants:    variants,
		Price:       feed.PricePromotions.Price,
		Available:   feed.AvailabilityInventory.Availability == "in_stock",
		Images:      images,
		Tags:        []string{feed.ItemInfo.Category, feed.ItemInfo.Brand}, // Extract tags from category and brand
	}
}

// LegacyProductToProductFeed converts a legacy Product to a ProductFeed (with minimal data)
func (ft *FeedTransformer) LegacyProductToProductFeed(product *Product) ProductFeed {
	// Create a basic ProductFeed from legacy Product data
	// This will have many fields empty as the legacy model doesn't have all the required fields
	return ProductFeed{
		OpenAIFlags: OpenAIFlags{
			EnableSearch:  true,
			EnableCheckout: false, // Default to false for legacy products
		},
		BasicData: BasicProductData{
			ID:          product.ID,
			Title:       product.Title,
			Description: product.Description,
			Link:        "", // Not available in legacy model
		},
		ItemInfo: ItemInformation{
			Category: "Legacy", // Default category
		},
		Media: Media{
			ImageLink:           "",
			AdditionalImageLink: product.Images,
		},
		PricePromotions: PricePromotions{
			Price: product.Price,
		},
		AvailabilityInventory: AvailabilityInventory{
			Availability:      "in_stock", // Default based on Available field
			InventoryQuantity: 1,          // Default quantity
		},
		MerchantInfo: MerchantInfo{
			SellerName: "Legacy Merchant",
			SellerURL:  "",
		},
		Returns: Returns{
			ReturnPolicy: "",
			ReturnWindow: 30, // Default return window
		},
	}
}

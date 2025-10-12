package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// ExampleProductFeed creates an example product feed that demonstrates all fields
func ExampleProductFeed() *ProductFeed {
	return &ProductFeed{
		OpenAIFlags: OpenAIFlags{
			EnableSearch:  true,
			EnableCheckout: true,
		},
		BasicData: BasicProductData{
			ID:          "SKU12345",
			GTIN:        "1234567890123",
			MPN:         "GPT5",
			Title:       "Men's Trail Running Shoes",
			Description: "Waterproof trail shoe with cushioned sole designed for rugged terrain and long-distance running.",
			Link:        "https://example.com/product/SKU12345",
		},
		ItemInfo: ItemInformation{
			Condition:  "new",
			Category:   "Apparel & Accessories > Shoes",
			Brand:      "OpenAI",
			Material:   "Synthetic Leather",
			Dimensions: "12x8x5 in",
			Weight:     "1.5 lb",
			AgeGroup:   "adult",
		},
		Media: Media{
			ImageLink:           "https://example.com/images/shoe-main.jpg",
			AdditionalImageLink: []string{"https://example.com/images/shoe-side.jpg", "https://example.com/images/shoe-back.jpg"},
			VideoLink:           "https://youtu.be/12345",
			Model3DLink:         "https://example.com/models/shoe.glb",
		},
		PricePromotions: PricePromotions{
			Price: Money{
				Value:    79.99,
				Currency: "USD",
			},
			SalePrice: &Money{
				Value:    59.99,
				Currency: "USD",
			},
			SalePriceEffectiveDate: "2025-07-01 / 2025-07-15",
			UnitPricingMeasure:     "1 pair",
			BaseMeasure:           "1 pair",
			PricingTrend:          "Lowest price in 6 months",
		},
		AvailabilityInventory: AvailabilityInventory{
			Availability:      "in_stock",
			InventoryQuantity: 25,
			PickupMethod:      "in_store",
			PickupSLA:         "1 day",
		},
		Variants: Variants{
			ItemGroupID:    "SHOE123GROUP",
			ItemGroupTitle: "Men's Trail Running Shoes",
			Color:          "Black",
			Size:           "10",
			SizeSystem:     "US",
			Gender:         "male",
			OfferID:        "SKU12345-Black-79.99",
		},
		Fulfillment: Fulfillment{
			Shipping: []string{
				"US:CA:Overnight:16.00 USD",
				"US:CA:Standard:8.00 USD",
				"US:CA:Economy:4.00 USD",
			},
			DeliveryEstimate: "2025-08-12",
		},
		MerchantInfo: MerchantInfo{
			SellerName:          "Example Store",
			SellerURL:           "https://example.com/store",
			SellerPrivacyPolicy: "https://example.com/privacy",
			SellerTOS:           "https://example.com/terms",
		},
		Returns: Returns{
			ReturnPolicy: "https://example.com/returns",
			ReturnWindow: 30,
		},
		PerformanceSignals: PerformanceSignals{
			PopularityScore: 4.7,
			ReturnRate:      2.0,
		},
		Compliance: Compliance{
			Warning:        "Contains lithium battery",
			WarningURL:     "https://example.com/warnings/battery",
			AgeRestriction: 0,
		},
		ReviewsQA: ReviewsQA{
			ProductReviewCount: 254,
			ProductReviewRating: 4.6,
			StoreReviewCount:   2000,
			StoreReviewRating:  4.8,
			QAndA:              "Q: Is this waterproof? A: Yes, these shoes are fully waterproof.",
			RawReviewData:      `{"reviews": [{"rating": 5, "comment": "Great shoes!"}]}`,
		},
		RelatedProducts: RelatedProducts{
			RelatedProductIDs: []string{"SKU67890", "SKU11111"},
			RelationshipType:  "often_bought_with",
		},
		GeoTagging: GeoTagging{
			GeoPrice: &Money{
				Value:    79.99,
				Currency: "USD",
			},
			GeoAvailability: "in_stock (California), out_of_stock (New York)",
		},
	}
}

// ExampleFeedIngestionRequest creates an example feed ingestion request
func ExampleFeedIngestionRequest() *FeedIngestionRequest {
	return &FeedIngestionRequest{
		MerchantID: "merchant_123",
		FeedData:   []ProductFeed{*ExampleProductFeed()},
		Format:     "json",
		Version:    "1.0",
		Metadata: map[string]interface{}{
			"source":       "api_upload",
			"processed_at": time.Now().UTC(),
			"file_size":    1024,
		},
	}
}

// ExampleCSVFeed creates an example CSV feed data
func ExampleCSVFeed() string {
	return `id,enable_search,enable_checkout,title,description,link,product_category,brand,price_value,price_currency,availability,inventory_quantity,seller_name,seller_url,return_policy,return_window
SKU12345,true,true,"Men's Trail Running Shoes","Waterproof trail shoe with cushioned sole","https://example.com/product/SKU12345","Apparel & Accessories > Shoes","OpenAI",79.99,USD,in_stock,25,"Example Store","https://example.com/store","https://example.com/returns",30
SKU67890,true,false,"Women's Running Shorts","Lightweight running shorts","https://example.com/product/SKU67890","Apparel & Accessories > Clothing","OpenAI",29.99,USD,in_stock,50,"Example Store","https://example.com/store","https://example.com/returns",30`
}

// ExampleFeedIngestionResponse creates an example successful feed ingestion response
func ExampleFeedIngestionResponse() *FeedIngestionResponse {
	return &FeedIngestionResponse{
		OK:            true,
		ProcessedCount: 1,
		ErrorCount:   0,
		FeedID:        "feed_12345",
		Message:       "Feed processed successfully",
	}
}

// ExampleFeedIngestionResponseWithErrors creates an example feed ingestion response with errors
func ExampleFeedIngestionResponseWithErrors() *FeedIngestionResponse {
	return &FeedIngestionResponse{
		OK:            false,
		ProcessedCount: 0,
		ErrorCount:    2,
		Errors: []string{
			"Product SKU12345: Missing required field 'title'",
			"Product SKU67890: Invalid currency code 'INVALID'",
		},
		FeedID:   "feed_12346",
		Message:  "Feed processing completed with errors",
	}
}

// ExampleFeedValidationResult creates an example validation result
func ExampleFeedValidationResult() *FeedValidationResult {
	return &FeedValidationResult{
		Valid:         true,
		Errors:        []ValidationError{},
		Warnings:      []ValidationWarning{
			{
				Field:     "brand",
				Message:   "Brand field is recommended for better discoverability",
				ProductID: "SKU12345",
				Code:      "MISSING_RECOMMENDED_FIELD",
			},
		},
		ProductCount:  1,
		ValidProducts: 1,
	}
}

// ExampleFeedStatus creates an example feed status
func ExampleFeedStatus() *FeedStatus {
	return &FeedStatus{
		FeedID:       "feed_12345",
		MerchantID:   "merchant_123",
		Status:       "completed",
		ProcessedAt:  time.Now().UTC(),
		ProductCount: 1,
		ErrorCount:   0,
		LastUpdated:  time.Now().UTC(),
	}
}

// ExampleProductSearchRequest creates an example product search request
func ExampleProductSearchRequest() *ProductSearchRequest {
	return &ProductSearchRequest{
		Query:        "running shoes",
		Category:     "Apparel & Accessories > Shoes",
		Brand:        "OpenAI",
		MinPrice:     &Money{Value: 50.0, Currency: "USD"},
		MaxPrice:     &Money{Value: 100.0, Currency: "USD"},
		Availability: "in_stock",
		MerchantID:   "merchant_123",
		Limit:        10,
		Offset:       0,
		SortBy:       "price",
		SortOrder:    "asc",
		Filters: map[string]interface{}{
			"color": "black",
			"size":  "10",
		},
	}
}

// ExampleProductSearchResponse creates an example product search response
func ExampleProductSearchResponse() *ProductSearchResponse {
	return &ProductSearchResponse{
		Products: []EnhancedProduct{
			{
				ProductFeed: *ExampleProductFeed(),
				CreatedAt:   time.Now().UTC().Add(-24 * time.Hour),
				UpdatedAt:   time.Now().UTC(),
				MerchantID:  "merchant_123",
				Status:      "active",
			},
		},
		Total:   1,
		Limit:   10,
		Offset:  0,
		HasMore: false,
	}
}

// ExampleValidationError creates an example validation error
func ExampleValidationError() ValidationError {
	return ValidationError{
		Field:     "title",
		Message:   "Title is required and cannot be empty",
		ProductID: "SKU12345",
		Severity:  "error",
		Code:      "REQUIRED_FIELD",
	}
}

// ExampleValidationWarning creates an example validation warning
func ExampleValidationWarning() ValidationWarning {
	return ValidationWarning{
		Field:     "brand",
		Message:   "Brand field is recommended for better product discoverability",
		ProductID: "SKU12345",
		Code:      "MISSING_RECOMMENDED_FIELD",
	}
}

// ExampleUsage demonstrates how to use the product feed models
func ExampleUsage() {
	fmt.Println("=== OpenAI Product Feed Specification - Go Models Example ===\n")

	// 1. Create a product feed
	productFeed := ExampleProductFeed()
	fmt.Println("1. Created Product Feed:")
	jsonData, _ := json.MarshalIndent(productFeed, "", "  ")
	fmt.Println(string(jsonData))
	fmt.Println()

	// 2. Validate the product feed
	validator := NewValidator()
	validationResult := validator.ValidateProductFeed(productFeed)
	fmt.Println("2. Validation Result:")
	fmt.Printf("Valid: %t\n", validationResult.Valid)
	fmt.Printf("Valid Products: %d/%d\n", validationResult.ValidProducts, validationResult.ProductCount)
	if len(validationResult.Errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range validationResult.Errors {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
	}
	if len(validationResult.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, warning := range validationResult.Warnings {
			fmt.Printf("  - %s: %s\n", warning.Field, warning.Message)
		}
	}
	fmt.Println()

	// 3. Create a feed ingestion request
	feedRequest := ExampleFeedIngestionRequest()
	fmt.Println("3. Feed Ingestion Request:")
	jsonData, _ = json.MarshalIndent(feedRequest, "", "  ")
	fmt.Println(string(jsonData))
	fmt.Println()

	// 4. Process the feed (simulate)
	processor := NewFeedProcessor()
	// In a real scenario, you would pass actual feed data
	fmt.Println("4. Feed Processing:")
	fmt.Println("Processing feed with format:", feedRequest.Format)
	fmt.Println("Number of products:", len(feedRequest.FeedData))
	fmt.Println()

	// 5. Transform to enhanced product
	transformer := NewFeedTransformer()
	enhancedProduct := transformer.ProductFeedToEnhancedProduct(*productFeed, "merchant_123")
	fmt.Println("5. Enhanced Product:")
	fmt.Printf("Product ID: %s\n", enhancedProduct.ProductFeed.BasicData.ID)
	fmt.Printf("Status: %s\n", enhancedProduct.Status)
	fmt.Printf("Merchant ID: %s\n", enhancedProduct.MerchantID)
	fmt.Printf("Created At: %s\n", enhancedProduct.CreatedAt.Format(time.RFC3339))
	fmt.Println()

	// 6. Transform to legacy product for backward compatibility
	legacyProduct := transformer.ProductFeedToLegacyProduct(*productFeed)
	fmt.Println("6. Legacy Product (for backward compatibility):")
	fmt.Printf("Product ID: %s\n", legacyProduct.ID)
	fmt.Printf("Title: %s\n", legacyProduct.Title)
	fmt.Printf("Available: %t\n", legacyProduct.Available)
	fmt.Printf("Price: %.2f %s\n", legacyProduct.Price.Value, legacyProduct.Price.Currency)
	fmt.Println()

	// 7. Store feed status
	storage := NewFeedStorage()
	feedStatus := ExampleFeedStatus()
	storage.StoreFeed(feedStatus.FeedID, feedStatus)
	fmt.Println("7. Feed Storage:")
	fmt.Printf("Stored feed: %s\n", feedStatus.FeedID)
	fmt.Printf("Status: %s\n", feedStatus.Status)
	fmt.Printf("Products processed: %d\n", feedStatus.ProductCount)
	fmt.Println()

	fmt.Println("=== Example completed successfully! ===")
}

// ExampleCSVProcessing demonstrates how to process CSV feeds
func ExampleCSVProcessing() {
	fmt.Println("=== CSV Feed Processing Example ===\n")

	csvData := ExampleCSVFeed()
	fmt.Println("1. CSV Feed Data:")
	fmt.Println(csvData)
	fmt.Println()

	processor := NewFeedProcessor()
	
	// Convert CSV data to bytes
	csvBytes := []byte(csvData)
	
	// Process the CSV feed
	feedRequest, err := processor.ProcessFeed("csv", csvBytes)
	if err != nil {
		fmt.Printf("Error processing CSV feed: %v\n", err)
		return
	}

	fmt.Println("2. Processed Feed Request:")
	fmt.Printf("Merchant ID: %s\n", feedRequest.MerchantID)
	fmt.Printf("Format: %s\n", feedRequest.Format)
	fmt.Printf("Number of products: %d\n", len(feedRequest.FeedData))
	fmt.Println()

	// Validate the processed feed
	validator := NewValidator()
	validationResult := validator.ValidateFeedIngestionRequest(feedRequest)
	
	fmt.Println("3. Validation Result:")
	fmt.Printf("Valid: %t\n", validationResult.Valid)
	fmt.Printf("Valid Products: %d/%d\n", validationResult.ValidProducts, validationResult.ProductCount)
	
	if len(validationResult.Errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range validationResult.Errors {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
	}
	
	if len(validationResult.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, warning := range validationResult.Warnings {
			fmt.Printf("  - %s: %s\n", warning.Field, warning.Message)
		}
	}
	
	fmt.Println("\n=== CSV Processing Example completed! ===")
}

package models

import (
	"encoding/json"
	"testing"
)

func TestProductFeedValidation(t *testing.T) {
	// Create a valid product feed
	productFeed := ExampleProductFeed()
	
	// Validate the feed
	validator := NewValidator()
	result := validator.ValidateProductFeed(productFeed)
	
	// Check if validation passes
	if !result.Valid {
		t.Errorf("Expected valid product feed, but got validation errors:")
		for _, err := range result.Errors {
			t.Errorf("  - %s: %s", err.Field, err.Message)
		}
	}
	
	// Check valid products count
	if result.ValidProducts != 1 {
		t.Errorf("Expected 1 valid product, got %d", result.ValidProducts)
	}
}

func TestProductFeedJSONSerialization(t *testing.T) {
	// Create a product feed
	productFeed := ExampleProductFeed()
	
	// Serialize to JSON
	jsonData, err := json.Marshal(productFeed)
	if err != nil {
		t.Fatalf("Failed to marshal product feed to JSON: %v", err)
	}
	
	// Deserialize from JSON
	var deserialized ProductFeed
	err = json.Unmarshal(jsonData, &deserialized)
	if err != nil {
		t.Fatalf("Failed to unmarshal product feed from JSON: %v", err)
	}
	
	// Verify key fields
	if deserialized.BasicData.ID != productFeed.BasicData.ID {
		t.Errorf("Expected ID %s, got %s", productFeed.BasicData.ID, deserialized.BasicData.ID)
	}
	
	if deserialized.BasicData.Title != productFeed.BasicData.Title {
		t.Errorf("Expected title %s, got %s", productFeed.BasicData.Title, deserialized.BasicData.Title)
	}
}

func TestFeedIngestionRequestValidation(t *testing.T) {
	// Create a valid feed ingestion request
	request := ExampleFeedIngestionRequest()
	
	// Validate the request
	validator := NewValidator()
	result := validator.ValidateFeedIngestionRequest(request)
	
	// Check if validation passes
	if !result.Valid {
		t.Errorf("Expected valid feed ingestion request, but got validation errors:")
		for _, err := range result.Errors {
			t.Errorf("  - %s: %s", err.Field, err.Message)
		}
	}
}

func TestFeedProcessorCSV(t *testing.T) {
	// Create CSV data
	csvData := ExampleCSVFeed()
	
	// Process the CSV feed
	processor := NewFeedProcessor()
	request, err := processor.ProcessFeed("csv", []byte(csvData))
	
	if err != nil {
		t.Fatalf("Failed to process CSV feed: %v", err)
	}
	
	// Check that we got products
	if len(request.FeedData) == 0 {
		t.Error("Expected at least one product from CSV feed")
	}
	
	// Check the first product
	product := request.FeedData[0]
	if product.BasicData.ID == "" {
		t.Error("Expected product ID to be populated")
	}
	
	if product.BasicData.Title == "" {
		t.Error("Expected product title to be populated")
	}
}

func TestFeedTransformer(t *testing.T) {
	// Create a product feed
	productFeed := ExampleProductFeed()
	
	// Transform to enhanced product
	transformer := NewFeedTransformer()
	enhancedProduct := transformer.ProductFeedToEnhancedProduct(*productFeed, "test_merchant")
	
	// Verify transformation
	if enhancedProduct.MerchantID != "test_merchant" {
		t.Errorf("Expected merchant ID 'test_merchant', got '%s'", enhancedProduct.MerchantID)
	}
	
	if enhancedProduct.ProductFeed.BasicData.ID != productFeed.BasicData.ID {
		t.Errorf("Expected product ID to match, got %s vs %s", 
			enhancedProduct.ProductFeed.BasicData.ID, productFeed.BasicData.ID)
	}
	
	// Transform to legacy product
	legacyProduct := transformer.ProductFeedToLegacyProduct(productFeed)
	
	// Verify legacy transformation
	if legacyProduct.ID != productFeed.BasicData.ID {
		t.Errorf("Expected legacy product ID to match, got %s vs %s", 
			legacyProduct.ID, productFeed.BasicData.ID)
	}
	
	if legacyProduct.Title != productFeed.BasicData.Title {
		t.Errorf("Expected legacy product title to match, got %s vs %s", 
			legacyProduct.Title, productFeed.BasicData.Title)
	}
}

func TestFeedStorage(t *testing.T) {
	// Create feed storage
	storage := NewFeedStorage()
	
	// Create a feed status
	status := ExampleFeedStatus()
	
	// Store the feed
	storage.StoreFeed(status.FeedID, status)
	
	// Retrieve the feed
	retrievedStatus, exists := storage.GetFeed(status.FeedID)
	if !exists {
		t.Error("Expected to find stored feed status")
	}
	
	if retrievedStatus.FeedID != status.FeedID {
		t.Errorf("Expected feed ID %s, got %s", status.FeedID, retrievedStatus.FeedID)
	}
	
	// Update the status
	err := storage.UpdateFeedStatus(status.FeedID, "completed", 100, 0)
	if err != nil {
		t.Errorf("Failed to update feed status: %v", err)
	}
	
	// Verify the update
	updatedStatus, _ := storage.GetFeed(status.FeedID)
	if updatedStatus.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", updatedStatus.Status)
	}
	
	if updatedStatus.ProductCount != 100 {
		t.Errorf("Expected product count 100, got %d", updatedStatus.ProductCount)
	}
}

func TestValidationErrors(t *testing.T) {
	// Create an invalid product feed (missing required fields)
	invalidFeed := &ProductFeed{
		BasicData: BasicProductData{
			// Missing ID, title, description, link
		},
		ItemInfo: ItemInformation{
			// Missing category
		},
		Media: Media{
			// Missing image_link
		},
		PricePromotions: PricePromotions{
			// Missing price
		},
		AvailabilityInventory: AvailabilityInventory{
			// Missing availability and inventory_quantity
		},
		MerchantInfo: MerchantInfo{
			// Missing seller_name and seller_url
		},
		Returns: Returns{
			// Missing return_policy and return_window
		},
	}
	
	// Validate the invalid feed
	validator := NewValidator()
	result := validator.ValidateProductFeed(invalidFeed)
	
	// Should not be valid
	if result.Valid {
		t.Error("Expected invalid product feed to fail validation")
	}
	
	// Should have multiple errors
	if len(result.Errors) == 0 {
		t.Error("Expected validation errors for invalid product feed")
	}
	
	// Check for specific required field errors
	hasIDError := false
	hasTitleError := false
	hasCategoryError := false
	
	for _, err := range result.Errors {
		if err.Field == "id" && err.Code == "REQUIRED_FIELD" {
			hasIDError = true
		}
		if err.Field == "title" && err.Code == "REQUIRED_FIELD" {
			hasTitleError = true
		}
		if err.Field == "product_category" && err.Code == "REQUIRED_FIELD" {
			hasCategoryError = true
		}
	}
	
	if !hasIDError {
		t.Error("Expected ID required field error")
	}
	if !hasTitleError {
		t.Error("Expected title required field error")
	}
	if !hasCategoryError {
		t.Error("Expected product_category required field error")
	}
}

# OpenAI Product Feed Specification - Go Models

This package provides comprehensive Go models for implementing the OpenAI Product Feed Specification within the ACP Gateway. These models enable merchants to submit structured product data that can be ingested, validated, and processed for use in ChatGPT's product discovery and checkout experiences.

## Overview

The OpenAI Product Feed Specification defines how merchants share structured product data with OpenAI so ChatGPT can accurately surface their products in search and shopping experiences. This implementation provides:

- **Complete field coverage**: All required, recommended, and optional fields from the specification
- **Comprehensive validation**: Built-in validation with detailed error reporting
- **Multiple format support**: JSON, CSV, TSV, and XML feed processing
- **Backward compatibility**: Integration with existing ACP Gateway models
- **Type safety**: Strong typing with proper Go struct definitions

## Key Components

### Core Models

#### ProductFeed
The main model representing a complete product feed entry according to the OpenAI specification:

```go
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
```

#### Field Categories

1. **OpenAI Flags**: Controls product visibility in ChatGPT
2. **Basic Product Data**: Core identifiers and descriptive text
3. **Item Information**: Physical characteristics and classification
4. **Media**: Visual and rich media assets
5. **Price & Promotions**: Standard and promotional pricing
6. **Availability & Inventory**: Stock levels and timing signals
7. **Variants**: Variant relationships and attributes
8. **Fulfillment**: Shipping methods and delivery times
9. **Merchant Info**: Seller identification and policies
10. **Returns**: Return policies and time windows
11. **Performance Signals**: Popularity and return-rate metrics
12. **Compliance**: Regulatory warnings and age restrictions
13. **Reviews and Q&A**: User-generated insights
14. **Related Products**: Cross-sell and substitute products
15. **Geo Tagging**: Region-specific pricing and availability

### Validation System

#### Validator
Comprehensive validation with detailed error reporting:

```go
validator := NewValidator()
result := validator.ValidateProductFeed(&productFeed)
if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("Error: %s - %s\n", err.Field, err.Message)
    }
}
```

#### Validation Features
- **Required field validation**: Ensures all mandatory fields are present
- **Format validation**: Validates URLs, dates, currencies, GTINs, etc.
- **Length constraints**: Enforces maximum character limits
- **Enum validation**: Validates allowed values for status fields
- **Dependency validation**: Ensures related fields are consistent
- **Warning system**: Flags missing recommended fields

### Feed Processing

#### FeedProcessor
Handles parsing and processing of different feed formats:

```go
processor := NewFeedProcessor()
request, err := processor.ProcessFeed("json", jsonData)
if err != nil {
    log.Fatal(err)
}
```

#### Supported Formats
- **JSON**: Native Go struct parsing
- **CSV**: Comma-separated values with header mapping
- **TSV**: Tab-separated values
- **XML**: XML structure parsing

### Data Transformation

#### FeedTransformer
Converts between different product representations:

```go
transformer := NewFeedTransformer()

// Convert to enhanced product
enhancedProduct := transformer.ProductFeedToEnhancedProduct(productFeed, merchantID)

// Convert to legacy product for backward compatibility
legacyProduct := transformer.ProductFeedToLegacyProduct(productFeed)
```

### Storage and Status Tracking

#### FeedStorage
Manages feed processing status and metadata:

```go
storage := NewFeedStorage()
status := &FeedStatus{
    FeedID:       "feed_123",
    MerchantID:   "merchant_456",
    Status:       "processing",
    ProductCount: 100,
    ErrorCount:   0,
}
storage.StoreFeed(status.FeedID, status)
```

## Usage Examples

### Basic Product Feed Creation

```go
package main

import (
    "fmt"
    "log"
    "your-project/internal/models"
)

func main() {
    // Create a product feed
    productFeed := &models.ProductFeed{
        OpenAIFlags: models.OpenAIFlags{
            EnableSearch:  true,
            EnableCheckout: true,
        },
        BasicData: models.BasicProductData{
            ID:          "SKU12345",
            GTIN:        "1234567890123",
            Title:       "Men's Trail Running Shoes",
            Description: "Waterproof trail shoe with cushioned sole",
            Link:        "https://example.com/product/SKU12345",
        },
        ItemInfo: models.ItemInformation{
            Category: "Apparel & Accessories > Shoes",
            Brand:    "OpenAI",
            Material: "Synthetic Leather",
            Weight:   "1.5 lb",
            AgeGroup: "adult",
        },
        Media: models.Media{
            ImageLink: "https://example.com/images/shoe-main.jpg",
        },
        PricePromotions: models.PricePromotions{
            Price: models.Money{
                Value:    79.99,
                Currency: "USD",
            },
        },
        AvailabilityInventory: models.AvailabilityInventory{
            Availability:      "in_stock",
            InventoryQuantity: 25,
        },
        MerchantInfo: models.MerchantInfo{
            SellerName: "Example Store",
            SellerURL:  "https://example.com/store",
        },
        Returns: models.Returns{
            ReturnPolicy: "https://example.com/returns",
            ReturnWindow: 30,
        },
    }

    // Validate the feed
    validator := models.NewValidator()
    result := validator.ValidateProductFeed(productFeed)
    
    if result.Valid {
        fmt.Println("Product feed is valid!")
    } else {
        fmt.Println("Validation errors:")
        for _, err := range result.Errors {
            fmt.Printf("- %s: %s\n", err.Field, err.Message)
        }
    }
}
```

### Feed Ingestion Request

```go
// Create a feed ingestion request
feedRequest := &models.FeedIngestionRequest{
    MerchantID: "merchant_123",
    FeedData:   []models.ProductFeed{*productFeed},
    Format:     "json",
    Version:    "1.0",
    Metadata: map[string]interface{}{
        "source": "api_upload",
        "file_size": 1024,
    },
}

// Validate the request
requestResult := validator.ValidateFeedIngestionRequest(feedRequest)
if !requestResult.Valid {
    log.Fatal("Feed ingestion request validation failed")
}
```

### CSV Feed Processing

```go
csvData := `id,title,description,link,product_category,brand,price_value,price_currency,availability,inventory_quantity,seller_name,seller_url,return_policy,return_window
SKU12345,"Men's Trail Running Shoes","Waterproof trail shoe","https://example.com/product/SKU12345","Apparel & Accessories > Shoes","OpenAI",79.99,USD,in_stock,25,"Example Store","https://example.com/store","https://example.com/returns",30`

processor := models.NewFeedProcessor()
request, err := processor.ProcessFeed("csv", []byte(csvData))
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Processed %d products from CSV feed\n", len(request.FeedData))
```

### Search and Retrieval

```go
// Create a search request
searchRequest := &models.ProductSearchRequest{
    Query:        "running shoes",
    Category:     "Apparel & Accessories > Shoes",
    MinPrice:     &models.Money{Value: 50.0, Currency: "USD"},
    MaxPrice:     &models.Money{Value: 100.0, Currency: "USD"},
    Availability: "in_stock",
    Limit:        10,
    SortBy:       "price",
    SortOrder:    "asc",
}

// Process search (implementation depends on your storage backend)
// searchResponse := yourSearchImplementation(searchRequest)
```

## Field Reference

### Required Fields

| Field | Type | Description | Max Length |
|-------|------|-------------|------------|
| `id` | string | Merchant product ID | 100 |
| `title` | string | Product title | 150 |
| `description` | string | Product description | 5000 |
| `link` | string | Product detail page URL | - |
| `product_category` | string | Category path | - |
| `price` | Money | Regular price | - |
| `availability` | string | Product availability | - |
| `inventory_quantity` | int | Stock count | - |
| `image_link` | string | Main product image URL | - |
| `seller_name` | string | Seller name | 70 |
| `seller_url` | string | Seller page URL | - |
| `return_policy` | string | Return policy URL | - |
| `return_window` | int | Days allowed for return | - |

### Recommended Fields

| Field | Type | Description | Max Length |
|-------|------|-------------|------------|
| `gtin` | string | Universal product identifier | 8-14 digits |
| `brand` | string | Product brand | 70 |
| `enable_search` | bool | Enable search in ChatGPT | - |
| `enable_checkout` | bool | Enable checkout in ChatGPT | - |
| `popularity_score` | float64 | Popularity indicator (0-5) | - |
| `return_rate` | float64 | Return rate percentage (0-100) | - |

### Optional Fields

| Field | Type | Description | Max Length |
|-------|------|-------------|------------|
| `mpn` | string | Manufacturer part number | 70 |
| `condition` | string | Product condition | - |
| `material` | string | Primary material(s) | 100 |
| `weight` | string | Product weight | - |
| `age_group` | string | Target demographic | - |
| `additional_image_link` | []string | Extra images | - |
| `video_link` | string | Product video URL | - |
| `sale_price` | Money | Discounted price | - |
| `color` | string | Variant color | 40 |
| `size` | string | Variant size | 20 |
| `gender` | string | Gender target | - |
| `shipping` | []string | Shipping methods | - |
| `warning` | string | Product disclaimers | 200 |
| `age_restriction` | int | Minimum purchase age | - |
| `product_review_count` | int | Number of reviews | - |
| `product_review_rating` | float64 | Average review score (0-5) | - |
| `related_product_id` | []string | Associated product IDs | - |

## Validation Rules

### OpenAI Flags
- `enable_checkout` requires `enable_search` to be true

### Basic Data
- `id` must be unique and stable over time
- `title` should avoid all-caps formatting
- `description` must be plain text only
- `link` must resolve with HTTP 200
- `gtin` must be 8, 12, 13, or 14 digits
- `mpn` is required if `gtin` is missing

### Pricing
- `price.value` must be greater than 0
- `price.currency` must be valid ISO 4217 code
- `sale_price` cannot exceed regular price
- `sale_price_effective_date` required if `sale_price` provided

### Availability
- `availability` must be: in_stock, out_of_stock, or preorder
- `availability_date` required for preorder items
- `inventory_quantity` must be non-negative

### Variants
- `item_group_id` required if variants exist
- `size_system` must be valid ISO 3166 country code
- `gender` must be: male, female, or unisex

### Merchant Info
- `seller_privacy_policy` required if `enable_checkout` is true
- `seller_tos` required if `enable_checkout` is true

### Performance Signals
- `popularity_score` must be 0-5 scale
- `return_rate` must be 0-100 percentage

### Compliance
- `age_restriction` must be positive integer
- `warning_url` must resolve with HTTP 200

## Error Handling

The validation system provides detailed error information:

```go
type ValidationError struct {
    Field       string `json:"field"`
    Message     string `json:"message"`
    ProductID   string `json:"product_id,omitempty"`
    Severity    string `json:"severity"`
    Code        string `json:"code"`
}
```

### Error Codes
- `REQUIRED_FIELD`: Missing required field
- `INVALID_FORMAT`: Invalid data format
- `INVALID_LENGTH`: Field exceeds maximum length
- `INVALID_VALUE`: Invalid field value
- `INVALID_URL`: Invalid URL format
- `INVALID_DATE`: Invalid date format
- `INVALID_GTIN`: Invalid GTIN format
- `INVALID_CURRENCY`: Invalid currency code
- `INVALID_AVAILABILITY`: Invalid availability status
- `INVALID_CONDITION`: Invalid product condition
- `INVALID_AGE_GROUP`: Invalid age group
- `INVALID_GENDER`: Invalid gender value
- `INVALID_RELATIONSHIP`: Invalid relationship type
- `INVALID_PICKUP_METHOD`: Invalid pickup method
- `INVALID_RATING`: Invalid rating value
- `INVALID_PERCENTAGE`: Invalid percentage value
- `INVALID_COUNTRY_CODE`: Invalid country code

## Integration with ACP Gateway

These models integrate seamlessly with the existing ACP Gateway architecture:

1. **API Endpoints**: Use with `/merchant/feed` endpoint for feed ingestion
2. **Storage**: Compatible with existing in-memory storage and can be extended for database storage
3. **Validation**: Integrates with existing validation middleware
4. **Search**: Works with existing product search functionality
5. **Backward Compatibility**: Maintains compatibility with legacy Product and ProductVariant models

## Best Practices

### Feed Creation
1. **Complete required fields**: Ensure all mandatory fields are populated
2. **Use recommended fields**: Include recommended fields for better discoverability
3. **Validate early**: Validate feeds before submission
4. **Maintain consistency**: Keep product IDs stable over time
5. **Update regularly**: Refresh feeds when products, pricing, or availability change

### Performance
1. **Batch processing**: Process feeds in batches for better performance
2. **Async validation**: Use background validation for large feeds
3. **Caching**: Cache validation results for repeated operations
4. **Error handling**: Implement proper error handling and retry logic

### Security
1. **Validate URLs**: Ensure all URLs are valid and accessible
2. **Sanitize input**: Clean user-provided data before processing
3. **Rate limiting**: Implement rate limiting for feed submissions
4. **Authentication**: Secure feed endpoints with proper authentication

## Testing

Use the provided example functions for testing:

```go
// Create example data
productFeed := models.ExampleProductFeed()
feedRequest := models.ExampleFeedIngestionRequest()
validationResult := models.ExampleFeedValidationResult()

// Run comprehensive example
models.ExampleUsage()

// Test CSV processing
models.ExampleCSVProcessing()
```

## Contributing

When adding new fields or validation rules:

1. Update the corresponding struct with proper JSON tags
2. Add validation logic to the appropriate validation function
3. Update the field reference documentation
4. Add example data for testing
5. Ensure backward compatibility with existing models

## License

This implementation follows the same license as the ACP Gateway project.

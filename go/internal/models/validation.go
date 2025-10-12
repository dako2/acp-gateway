package models

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Validation constants
const (
	MaxProductIDLength     = 100
	MaxTitleLength         = 150
	MaxDescriptionLength   = 5000
	MaxBrandLength         = 70
	MaxMPNLength           = 70
	MaxMaterialLength      = 100
	MaxColorLength         = 40
	MaxSizeLength          = 20
	MaxItemGroupIDLength   = 70
	MaxItemGroupTitleLength = 150
	MaxPricingTrendLength  = 80
	MaxSellerNameLength    = 70
	MaxWarningLength       = 200
)

// ValidationErrorCode represents error codes for validation
type ValidationErrorCode string

const (
	ErrRequiredField       ValidationErrorCode = "REQUIRED_FIELD"
	ErrInvalidFormat       ValidationErrorCode = "INVALID_FORMAT"
	ErrInvalidLength       ValidationErrorCode = "INVALID_LENGTH"
	ErrInvalidValue        ValidationErrorCode = "INVALID_VALUE"
	ErrInvalidURL          ValidationErrorCode = "INVALID_URL"
	ErrInvalidDate         ValidationErrorCode = "INVALID_DATE"
	ErrInvalidGTIN         ValidationErrorCode = "INVALID_GTIN"
	ErrInvalidCurrency     ValidationErrorCode = "INVALID_CURRENCY"
	ErrInvalidAvailability ValidationErrorCode = "INVALID_AVAILABILITY"
	ErrInvalidCondition    ValidationErrorCode = "INVALID_CONDITION"
	ErrInvalidAgeGroup     ValidationErrorCode = "INVALID_AGE_GROUP"
	ErrInvalidGender       ValidationErrorCode = "INVALID_GENDER"
	ErrInvalidRelationship ValidationErrorCode = "INVALID_RELATIONSHIP"
	ErrInvalidPickupMethod ValidationErrorCode = "INVALID_PICKUP_METHOD"
	ErrInvalidRating       ValidationErrorCode = "INVALID_RATING"
	ErrInvalidPercentage   ValidationErrorCode = "INVALID_PERCENTAGE"
	ErrInvalidCountryCode  ValidationErrorCode = "INVALID_COUNTRY_CODE"
)

// Validator provides validation functions for product feed data
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateProductFeed validates a complete product feed entry
func (v *Validator) ValidateProductFeed(feed *ProductFeed) *FeedValidationResult {
	result := &FeedValidationResult{
		Valid:        true,
		Errors:       []ValidationError{},
		Warnings:     []ValidationWarning{},
		ProductCount: 1,
		ValidProducts: 0,
	}

	// Validate OpenAI flags
	v.validateOpenAIFlags(feed.OpenAIFlags, feed.BasicData.ID, result)

	// Validate basic product data
	v.validateBasicProductData(feed.BasicData, result)

	// Validate item information
	v.validateItemInformation(feed.ItemInfo, feed.BasicData.ID, result)

	// Validate media
	v.validateMedia(feed.Media, feed.BasicData.ID, result)

	// Validate price and promotions
	v.validatePricePromotions(feed.PricePromotions, feed.BasicData.ID, result)

	// Validate availability and inventory
	v.validateAvailabilityInventory(feed.AvailabilityInventory, feed.BasicData.ID, result)

	// Validate variants
	v.validateVariants(feed.Variants, feed.BasicData.ID, result)

	// Validate fulfillment
	v.validateFulfillment(feed.Fulfillment, feed.BasicData.ID, result)

	// Validate merchant info
	v.validateMerchantInfo(feed.MerchantInfo, feed.BasicData.ID, result)

	// Validate returns
	v.validateReturns(feed.Returns, feed.BasicData.ID, result)

	// Validate performance signals (optional)
	if feed.PerformanceSignals.PopularityScore != 0 || feed.PerformanceSignals.ReturnRate != 0 {
		v.validatePerformanceSignals(feed.PerformanceSignals, feed.BasicData.ID, result)
	}

	// Validate compliance (optional)
	if feed.Compliance.Warning != "" || feed.Compliance.WarningURL != "" || feed.Compliance.AgeRestriction != 0 {
		v.validateCompliance(feed.Compliance, feed.BasicData.ID, result)
	}

	// Validate reviews and Q&A (optional)
	if feed.ReviewsQA.ProductReviewCount != 0 || feed.ReviewsQA.ProductReviewRating != 0 || feed.ReviewsQA.QAndA != "" {
		v.validateReviewsQA(feed.ReviewsQA, feed.BasicData.ID, result)
	}

	// Validate related products (optional)
	if len(feed.RelatedProducts.RelatedProductIDs) > 0 {
		v.validateRelatedProducts(feed.RelatedProducts, feed.BasicData.ID, result)
	}

	// Validate geo tagging (optional)
	if feed.GeoTagging.GeoPrice != nil || feed.GeoTagging.GeoAvailability != "" {
		v.validateGeoTagging(feed.GeoTagging, feed.BasicData.ID, result)
	}

	// Determine overall validity
	result.Valid = len(result.Errors) == 0
	if result.Valid {
		result.ValidProducts = 1
	}

	return result
}

// validateOpenAIFlags validates OpenAI-specific flags
func (v *Validator) validateOpenAIFlags(flags OpenAIFlags, productID string, result *FeedValidationResult) {
	if !flags.EnableSearch && flags.EnableCheckout {
		v.addError(result, "enable_checkout", "enable_checkout requires enable_search to be true", productID, ErrInvalidValue)
	}
}

// validateBasicProductData validates basic product information
func (v *Validator) validateBasicProductData(data BasicProductData, result *FeedValidationResult) {
	if data.ID == "" {
		v.addError(result, "id", "product ID is required", data.ID, ErrRequiredField)
	} else if len(data.ID) > MaxProductIDLength {
		v.addError(result, "id", fmt.Sprintf("product ID must be %d characters or less", MaxProductIDLength), data.ID, ErrInvalidLength)
	}

	if data.Title == "" {
		v.addError(result, "title", "title is required", data.ID, ErrRequiredField)
	} else if len(data.Title) > MaxTitleLength {
		v.addError(result, "title", fmt.Sprintf("title must be %d characters or less", MaxTitleLength), data.ID, ErrInvalidLength)
	}

	if data.Description == "" {
		v.addError(result, "description", "description is required", data.ID, ErrRequiredField)
	} else if len(data.Description) > MaxDescriptionLength {
		v.addError(result, "description", fmt.Sprintf("description must be %d characters or less", MaxDescriptionLength), data.ID, ErrInvalidLength)
	}

	if data.Link == "" {
		v.addError(result, "link", "product link is required", data.ID, ErrRequiredField)
	} else if !v.isValidURL(data.Link) {
		v.addError(result, "link", "invalid URL format", data.ID, ErrInvalidURL)
	}

	// Validate GTIN if provided
	if data.GTIN != "" {
		if !v.isValidGTIN(data.GTIN) {
			v.addError(result, "gtin", "invalid GTIN format (must be 8, 12, 13, or 14 digits)", data.ID, ErrInvalidGTIN)
		}
	}

	// Validate MPN if GTIN is not provided
	if data.GTIN == "" && data.MPN == "" {
		v.addError(result, "mpn", "MPN is required when GTIN is not provided", data.ID, ErrRequiredField)
	} else if data.MPN != "" && len(data.MPN) > MaxMPNLength {
		v.addError(result, "mpn", fmt.Sprintf("MPN must be %d characters or less", MaxMPNLength), data.ID, ErrInvalidLength)
	}
}

// validateItemInformation validates item-specific information
func (v *Validator) validateItemInformation(info ItemInformation, productID string, result *FeedValidationResult) {
	if info.Category == "" {
		v.addError(result, "product_category", "product category is required", productID, ErrRequiredField)
	}

	if info.Brand != "" && len(info.Brand) > MaxBrandLength {
		v.addError(result, "brand", fmt.Sprintf("brand must be %d characters or less", MaxBrandLength), productID, ErrInvalidLength)
	}

	if info.Material != "" && len(info.Material) > MaxMaterialLength {
		v.addError(result, "material", fmt.Sprintf("material must be %d characters or less", MaxMaterialLength), productID, ErrInvalidLength)
	}

	if info.Condition != "" && !v.isValidCondition(info.Condition) {
		v.addError(result, "condition", "invalid condition (must be new, refurbished, or used)", productID, ErrInvalidCondition)
	}

	if info.AgeGroup != "" && !v.isValidAgeGroup(info.AgeGroup) {
		v.addError(result, "age_group", "invalid age group (must be newborn, infant, toddler, kids, or adult)", productID, ErrInvalidAgeGroup)
	}
}

// validateMedia validates media assets
func (v *Validator) validateMedia(media Media, productID string, result *FeedValidationResult) {
	if media.ImageLink == "" {
		v.addError(result, "image_link", "main image link is required", productID, ErrRequiredField)
	} else if !v.isValidURL(media.ImageLink) {
		v.addError(result, "image_link", "invalid image URL format", productID, ErrInvalidURL)
	}

	for i, link := range media.AdditionalImageLink {
		if !v.isValidURL(link) {
			v.addError(result, fmt.Sprintf("additional_image_link[%d]", i), "invalid additional image URL format", productID, ErrInvalidURL)
		}
	}

	if media.VideoLink != "" && !v.isValidURL(media.VideoLink) {
		v.addError(result, "video_link", "invalid video URL format", productID, ErrInvalidURL)
	}

	if media.Model3DLink != "" && !v.isValidURL(media.Model3DLink) {
		v.addError(result, "model_3d_link", "invalid 3D model URL format", productID, ErrInvalidURL)
	}
}

// validatePricePromotions validates pricing information
func (v *Validator) validatePricePromotions(pricing PricePromotions, productID string, result *FeedValidationResult) {
	if pricing.Price.Value <= 0 {
		v.addError(result, "price.value", "price value must be greater than 0", productID, ErrInvalidValue)
	}

	if pricing.Price.Currency == "" {
		v.addError(result, "price.currency", "currency is required", productID, ErrRequiredField)
	} else if !v.isValidCurrency(pricing.Price.Currency) {
		v.addError(result, "price.currency", "invalid currency code", productID, ErrInvalidCurrency)
	}

	if pricing.SalePrice != nil {
		if pricing.SalePrice.Value > pricing.Price.Value {
			v.addError(result, "sale_price.value", "sale price cannot be greater than regular price", productID, ErrInvalidValue)
		}

		if pricing.SalePrice.Currency != pricing.Price.Currency {
			v.addError(result, "sale_price.currency", "sale price currency must match regular price currency", productID, ErrInvalidValue)
		}

		if pricing.SalePriceEffectiveDate == "" {
			v.addError(result, "sale_price_effective_date", "sale price effective date is required when sale price is provided", productID, ErrRequiredField)
		} else if !v.isValidDateRange(pricing.SalePriceEffectiveDate) {
			v.addError(result, "sale_price_effective_date", "invalid date range format", productID, ErrInvalidDate)
		}
	}

	if pricing.PricingTrend != "" && len(pricing.PricingTrend) > MaxPricingTrendLength {
		v.addError(result, "pricing_trend", fmt.Sprintf("pricing trend must be %d characters or less", MaxPricingTrendLength), productID, ErrInvalidLength)
	}
}

// validateAvailabilityInventory validates availability and inventory data
func (v *Validator) validateAvailabilityInventory(availability AvailabilityInventory, productID string, result *FeedValidationResult) {
	if !v.isValidAvailability(availability.Availability) {
		v.addError(result, "availability", "invalid availability status (must be in_stock, out_of_stock, or preorder)", productID, ErrInvalidAvailability)
	}

	if availability.Availability == "preorder" && availability.AvailabilityDate == "" {
		v.addError(result, "availability_date", "availability date is required for preorder items", productID, ErrRequiredField)
	} else if availability.AvailabilityDate != "" && !v.isValidDate(availability.AvailabilityDate) {
		v.addError(result, "availability_date", "invalid availability date format", productID, ErrInvalidDate)
	}

	if availability.InventoryQuantity < 0 {
		v.addError(result, "inventory_quantity", "inventory quantity must be non-negative", productID, ErrInvalidValue)
	}

	if availability.ExpirationDate != "" && !v.isValidDate(availability.ExpirationDate) {
		v.addError(result, "expiration_date", "invalid expiration date format", productID, ErrInvalidDate)
	}

	if availability.PickupMethod != "" && !v.isValidPickupMethod(availability.PickupMethod) {
		v.addError(result, "pickup_method", "invalid pickup method (must be in_store, reserve, or not_supported)", productID, ErrInvalidPickupMethod)
	}
}

// validateVariants validates variant information
func (v *Validator) validateVariants(variants Variants, productID string, result *FeedValidationResult) {
	if variants.ItemGroupID != "" && len(variants.ItemGroupID) > MaxItemGroupIDLength {
		v.addError(result, "item_group_id", fmt.Sprintf("item group ID must be %d characters or less", MaxItemGroupIDLength), productID, ErrInvalidLength)
	}

	if variants.ItemGroupTitle != "" && len(variants.ItemGroupTitle) > MaxItemGroupTitleLength {
		v.addError(result, "item_group_title", fmt.Sprintf("item group title must be %d characters or less", MaxItemGroupTitleLength), productID, ErrInvalidLength)
	}

	if variants.Color != "" && len(variants.Color) > MaxColorLength {
		v.addError(result, "color", fmt.Sprintf("color must be %d characters or less", MaxColorLength), productID, ErrInvalidLength)
	}

	if variants.Size != "" && len(variants.Size) > MaxSizeLength {
		v.addError(result, "size", fmt.Sprintf("size must be %d characters or less", MaxSizeLength), productID, ErrInvalidLength)
	}

	if variants.SizeSystem != "" && !v.isValidCountryCode(variants.SizeSystem) {
		v.addError(result, "size_system", "invalid size system country code", productID, ErrInvalidCountryCode)
	}

	if variants.Gender != "" && !v.isValidGender(variants.Gender) {
		v.addError(result, "gender", "invalid gender (must be male, female, or unisex)", productID, ErrInvalidGender)
	}
}

// validateFulfillment validates fulfillment information
func (v *Validator) validateFulfillment(fulfillment Fulfillment, productID string, result *FeedValidationResult) {
	for i, shipping := range fulfillment.Shipping {
		if !v.isValidShippingFormat(shipping) {
			v.addError(result, fmt.Sprintf("shipping[%d]", i), "invalid shipping format", productID, ErrInvalidFormat)
		}
	}

	if fulfillment.DeliveryEstimate != "" && !v.isValidDate(fulfillment.DeliveryEstimate) {
		v.addError(result, "delivery_estimate", "invalid delivery estimate date format", productID, ErrInvalidDate)
	}
}

// validateMerchantInfo validates merchant information
func (v *Validator) validateMerchantInfo(merchant MerchantInfo, productID string, result *FeedValidationResult) {
	if merchant.SellerName == "" {
		v.addError(result, "seller_name", "seller name is required", productID, ErrRequiredField)
	} else if len(merchant.SellerName) > MaxSellerNameLength {
		v.addError(result, "seller_name", fmt.Sprintf("seller name must be %d characters or less", MaxSellerNameLength), productID, ErrInvalidLength)
	}

	if merchant.SellerURL == "" {
		v.addError(result, "seller_url", "seller URL is required", productID, ErrRequiredField)
	} else if !v.isValidURL(merchant.SellerURL) {
		v.addError(result, "seller_url", "invalid seller URL format", productID, ErrInvalidURL)
	}

	if merchant.SellerPrivacyPolicy != "" && !v.isValidURL(merchant.SellerPrivacyPolicy) {
		v.addError(result, "seller_privacy_policy", "invalid privacy policy URL format", productID, ErrInvalidURL)
	}

	if merchant.SellerTOS != "" && !v.isValidURL(merchant.SellerTOS) {
		v.addError(result, "seller_tos", "invalid terms of service URL format", productID, ErrInvalidURL)
	}
}

// validateReturns validates return policy information
func (v *Validator) validateReturns(returns Returns, productID string, result *FeedValidationResult) {
	if returns.ReturnPolicy == "" {
		v.addError(result, "return_policy", "return policy URL is required", productID, ErrRequiredField)
	} else if !v.isValidURL(returns.ReturnPolicy) {
		v.addError(result, "return_policy", "invalid return policy URL format", productID, ErrInvalidURL)
	}

	if returns.ReturnWindow <= 0 {
		v.addError(result, "return_window", "return window must be a positive integer", productID, ErrInvalidValue)
	}
}

// validatePerformanceSignals validates performance metrics
func (v *Validator) validatePerformanceSignals(signals PerformanceSignals, productID string, result *FeedValidationResult) {
	if signals.PopularityScore < 0 || signals.PopularityScore > 5 {
		v.addError(result, "popularity_score", "popularity score must be between 0 and 5", productID, ErrInvalidRating)
	}

	if signals.ReturnRate < 0 || signals.ReturnRate > 100 {
		v.addError(result, "return_rate", "return rate must be between 0 and 100", productID, ErrInvalidPercentage)
	}
}

// validateCompliance validates compliance information
func (v *Validator) validateCompliance(compliance Compliance, productID string, result *FeedValidationResult) {
	if compliance.Warning != "" && len(compliance.Warning) > MaxWarningLength {
		v.addError(result, "warning", fmt.Sprintf("warning must be %d characters or less", MaxWarningLength), productID, ErrInvalidLength)
	}

	if compliance.WarningURL != "" && !v.isValidURL(compliance.WarningURL) {
		v.addError(result, "warning_url", "invalid warning URL format", productID, ErrInvalidURL)
	}

	if compliance.AgeRestriction < 0 {
		v.addError(result, "age_restriction", "age restriction must be a positive integer", productID, ErrInvalidValue)
	}
}

// validateReviewsQA validates reviews and Q&A information
func (v *Validator) validateReviewsQA(reviews ReviewsQA, productID string, result *FeedValidationResult) {
	if reviews.ProductReviewCount < 0 {
		v.addError(result, "product_review_count", "product review count must be non-negative", productID, ErrInvalidValue)
	}

	if reviews.ProductReviewRating < 0 || reviews.ProductReviewRating > 5 {
		v.addError(result, "product_review_rating", "product review rating must be between 0 and 5", productID, ErrInvalidRating)
	}

	if reviews.StoreReviewCount < 0 {
		v.addError(result, "store_review_count", "store review count must be non-negative", productID, ErrInvalidValue)
	}

	if reviews.StoreReviewRating < 0 || reviews.StoreReviewRating > 5 {
		v.addError(result, "store_review_rating", "store review rating must be between 0 and 5", productID, ErrInvalidRating)
	}
}

// validateRelatedProducts validates related products information
func (v *Validator) validateRelatedProducts(related RelatedProducts, productID string, result *FeedValidationResult) {
	if related.RelationshipType != "" && !v.isValidRelationshipType(related.RelationshipType) {
		v.addError(result, "relationship_type", "invalid relationship type", productID, ErrInvalidRelationship)
	}
}

// validateGeoTagging validates geo-specific information
func (v *Validator) validateGeoTagging(geo GeoTagging, productID string, result *FeedValidationResult) {
	if geo.GeoPrice != nil {
		if geo.GeoPrice.Value <= 0 {
			v.addError(result, "geo_price.value", "geo price value must be greater than 0", productID, ErrInvalidValue)
		}

		if geo.GeoPrice.Currency == "" {
			v.addError(result, "geo_price.currency", "geo price currency is required", productID, ErrRequiredField)
		} else if !v.isValidCurrency(geo.GeoPrice.Currency) {
			v.addError(result, "geo_price.currency", "invalid geo price currency code", productID, ErrInvalidCurrency)
		}
	}
}

// Helper validation functions

func (v *Validator) isValidURL(urlStr string) bool {
	_, err := url.ParseRequestURI(urlStr)
	return err == nil
}

func (v *Validator) isValidGTIN(gtin string) bool {
	matched, _ := regexp.MatchString(`^\d{8}$|^\d{12}$|^\d{13}$|^\d{14}$`, gtin)
	return matched
}

func (v *Validator) isValidCurrency(currency string) bool {
	validCurrencies := []string{"USD", "EUR", "GBP", "CAD", "AUD", "JPY", "CHF", "CNY", "INR", "BRL", "MXN", "KRW", "SGD", "HKD", "NOK", "SEK", "DKK", "PLN", "CZK", "HUF", "ILS", "CLP", "PHP", "AED", "SAR", "ZAR", "THB", "MYR", "IDR", "VND", "TRY", "RUB", "UAH", "KZT", "BGN", "RON", "HRK", "ISK", "NZD", "EGP", "QAR", "KWD", "BHD", "OMR", "JOD", "LBP", "PKR", "BDT", "LKR", "NPR", "AFN", "AMD", "AZN", "GEL", "KGS", "TJS", "TMT", "UZS"}
	for _, valid := range validCurrencies {
		if currency == valid {
			return true
		}
	}
	return false
}

func (v *Validator) isValidAvailability(availability string) bool {
	validStatuses := []string{"in_stock", "out_of_stock", "preorder"}
	for _, status := range validStatuses {
		if availability == status {
			return true
		}
	}
	return false
}

func (v *Validator) isValidCondition(condition string) bool {
	validConditions := []string{"new", "refurbished", "used"}
	for _, valid := range validConditions {
		if condition == valid {
			return true
		}
	}
	return false
}

func (v *Validator) isValidAgeGroup(ageGroup string) bool {
	validAgeGroups := []string{"newborn", "infant", "toddler", "kids", "adult"}
	for _, valid := range validAgeGroups {
		if ageGroup == valid {
			return true
		}
	}
	return false
}

func (v *Validator) isValidGender(gender string) bool {
	validGenders := []string{"male", "female", "unisex"}
	for _, valid := range validGenders {
		if gender == valid {
			return true
		}
	}
	return false
}

func (v *Validator) isValidPickupMethod(method string) bool {
	validMethods := []string{"in_store", "reserve", "not_supported"}
	for _, valid := range validMethods {
		if method == valid {
			return true
		}
	}
	return false
}

func (v *Validator) isValidRelationshipType(relationship string) bool {
	validTypes := []string{"part_of_set", "required_part", "often_bought_with", "substitute", "different_brand", "accessory"}
	for _, valid := range validTypes {
		if relationship == valid {
			return true
		}
	}
	return false
}

func (v *Validator) isValidCountryCode(code string) bool {
	// Simplified country code validation - in production, use a proper country code library
	matched, _ := regexp.MatchString(`^[A-Z]{2}$`, code)
	return matched
}

func (v *Validator) isValidDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

func (v *Validator) isValidDateRange(dateRange string) bool {
	parts := strings.Split(dateRange, " / ")
	if len(parts) != 2 {
		return false
	}
	startDate, err1 := time.Parse("2006-01-02", strings.TrimSpace(parts[0]))
	endDate, err2 := time.Parse("2006-01-02", strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return false
	}
	return startDate.Before(endDate)
}

func (v *Validator) isValidShippingFormat(shipping string) bool {
	// Format: country:region:service_class:price
	// Example: "US:CA:Overnight:16.00 USD"
	parts := strings.Split(shipping, ":")
	return len(parts) == 4
}

// Helper functions for adding validation results

func (v *Validator) addError(result *FeedValidationResult, field, message, productID string, code ValidationErrorCode) {
	result.Errors = append(result.Errors, ValidationError{
		Field:     field,
		Message:   message,
		ProductID: productID,
		Severity:  "error",
		Code:      string(code),
	})
}

func (v *Validator) addWarning(result *FeedValidationResult, field, message, productID string, code ValidationErrorCode) {
	result.Warnings = append(result.Warnings, ValidationWarning{
		Field:     field,
		Message:   message,
		ProductID: productID,
		Code:      string(code),
	})
}

// ValidateFeedIngestionRequest validates a feed ingestion request
func (v *Validator) ValidateFeedIngestionRequest(request *FeedIngestionRequest) *FeedValidationResult {
	result := &FeedValidationResult{
		Valid:         true,
		Errors:        []ValidationError{},
		Warnings:      []ValidationWarning{},
		ProductCount:  len(request.FeedData),
		ValidProducts: 0,
	}

	if request.MerchantID == "" {
		v.addError(result, "merchant_id", "merchant ID is required", "", ErrRequiredField)
	}

	if len(request.FeedData) == 0 {
		v.addError(result, "feed_data", "feed data is required", "", ErrRequiredField)
	}

	validFormats := []string{"json", "csv", "tsv", "xml"}
	isValidFormat := false
	for _, format := range validFormats {
		if request.Format == format {
			isValidFormat = true
			break
		}
	}
	if !isValidFormat {
		v.addError(result, "format", "invalid format (must be json, csv, tsv, or xml)", "", ErrInvalidFormat)
	}

	// Validate each product in the feed
	for _, product := range request.FeedData {
		productResult := v.ValidateProductFeed(&product)
		result.Errors = append(result.Errors, productResult.Errors...)
		result.Warnings = append(result.Warnings, productResult.Warnings...)
		if productResult.Valid {
			result.ValidProducts++
		}
	}

	result.Valid = len(result.Errors) == 0
	return result
}

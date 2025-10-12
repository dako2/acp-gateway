"""
OpenAI Product Feed Specification Models for FastAPI

This module provides comprehensive Pydantic models for implementing the OpenAI Product Feed Specification
within the ACP Gateway FastAPI implementation.
"""

from pydantic import BaseModel, Field, validator, HttpUrl
from typing import List, Optional, Dict, Any, Literal, Union
from datetime import datetime, date
from enum import Enum
import re


# ==================== ENUMS ====================

class AvailabilityStatus(str, Enum):
    IN_STOCK = "in_stock"
    OUT_OF_STOCK = "out_of_stock"
    PREORDER = "preorder"


class ProductCondition(str, Enum):
    NEW = "new"
    REFURBISHED = "refurbished"
    USED = "used"


class AgeGroup(str, Enum):
    NEWBORN = "newborn"
    INFANT = "infant"
    TODDLER = "toddler"
    KIDS = "kids"
    ADULT = "adult"


class Gender(str, Enum):
    MALE = "male"
    FEMALE = "female"
    UNISEX = "unisex"


class PickupMethod(str, Enum):
    IN_STORE = "in_store"
    RESERVE = "reserve"
    NOT_SUPPORTED = "not_supported"


class RelationshipType(str, Enum):
    PART_OF_SET = "part_of_set"
    REQUIRED_PART = "required_part"
    OFTEN_BOUGHT_WITH = "often_bought_with"
    SUBSTITUTE = "substitute"
    DIFFERENT_BRAND = "different_brand"
    ACCESSORY = "accessory"


# ==================== CORE MODELS ====================

class Money(BaseModel):
    """Monetary amount with currency"""
    value: float = Field(..., gt=0, description="Monetary value")
    currency: str = Field(..., min_length=3, max_length=3, description="ISO 4217 currency code")

    @validator('currency')
    def validate_currency(cls, v):
        valid_currencies = {
            'USD', 'EUR', 'GBP', 'CAD', 'AUD', 'JPY', 'CHF', 'CNY', 'INR', 'BRL', 'MXN', 
            'KRW', 'SGD', 'HKD', 'NOK', 'SEK', 'DKK', 'PLN', 'CZK', 'HUF', 'ILS', 'CLP', 
            'PHP', 'AED', 'SAR', 'ZAR', 'THB', 'MYR', 'IDR', 'VND', 'TRY', 'RUB', 'UAH', 
            'KZT', 'BGN', 'RON', 'HRK', 'ISK', 'NZD', 'EGP', 'QAR', 'KWD', 'BHD', 'OMR', 
            'JOD', 'LBP', 'PKR', 'BDT', 'LKR', 'NPR', 'AFN', 'AMD', 'AZN', 'GEL', 'KGS', 
            'TJS', 'TMT', 'UZS'
        }
        if v.upper() not in valid_currencies:
            raise ValueError(f'Invalid currency code: {v}')
        return v.upper()


# ==================== PRODUCT FEED MODELS ====================

class OpenAIFlags(BaseModel):
    """OpenAI-specific flags controlling product visibility in ChatGPT"""
    enable_search: bool = Field(..., description="Controls whether product can be surfaced in ChatGPT search results")
    enable_checkout: bool = Field(..., description="Allows direct purchase inside ChatGPT")

    @validator('enable_checkout')
    def validate_checkout_requires_search(cls, v, values):
        if v and not values.get('enable_search'):
            raise ValueError('enable_checkout requires enable_search to be true')
        return v


class BasicProductData(BaseModel):
    """Core identifiers and descriptive text for the product"""
    id: str = Field(..., max_length=100, description="Merchant product ID (unique)")
    gtin: Optional[str] = Field(None, description="Universal product identifier")
    mpn: Optional[str] = Field(None, max_length=70, description="Manufacturer part number")
    title: str = Field(..., max_length=150, description="Product title")
    description: str = Field(..., max_length=5000, description="Full product description")
    link: HttpUrl = Field(..., description="Product detail page URL")

    @validator('gtin')
    def validate_gtin(cls, v):
        if v and not re.match(r'^\d{8}$|^\d{12}$|^\d{13}$|^\d{14}$', v):
            raise ValueError('GTIN must be 8, 12, 13, or 14 digits')
        return v

    @validator('mpn')
    def validate_mpn_required_if_no_gtin(cls, v, values):
        if not values.get('gtin') and not v:
            raise ValueError('MPN is required when GTIN is not provided')
        return v


class ItemInformation(BaseModel):
    """Physical characteristics and classification details"""
    condition: Optional[ProductCondition] = Field(None, description="Condition of product")
    product_category: str = Field(..., description="Category path")
    brand: Optional[str] = Field(None, max_length=70, description="Product brand")
    material: Optional[str] = Field(None, max_length=100, description="Primary material(s)")
    dimensions: Optional[str] = Field(None, description="Overall dimensions")
    length: Optional[str] = Field(None, description="Individual dimension")
    width: Optional[str] = Field(None, description="Individual dimension")
    height: Optional[str] = Field(None, description="Individual dimension")
    weight: Optional[str] = Field(None, description="Product weight")
    age_group: Optional[AgeGroup] = Field(None, description="Target demographic")


class Media(BaseModel):
    """Visual and rich media assets"""
    image_link: HttpUrl = Field(..., description="Main product image URL")
    additional_image_link: Optional[List[HttpUrl]] = Field(None, description="Extra images")
    video_link: Optional[HttpUrl] = Field(None, description="Product video URL")
    model_3d_link: Optional[HttpUrl] = Field(None, description="3D model URL")


class PricePromotions(BaseModel):
    """Standard and promotional pricing information"""
    price: Money = Field(..., description="Regular price")
    applicable_taxes_fees: Optional[Money] = Field(None, description="Additional taxes/fees")
    sale_price: Optional[Money] = Field(None, description="Discounted price")
    sale_price_effective_date: Optional[str] = Field(None, description="Sale window (ISO 8601 date range)")
    unit_pricing_measure: Optional[str] = Field(None, description="Unit price measure")
    base_measure: Optional[str] = Field(None, description="Base measure")
    pricing_trend: Optional[str] = Field(None, max_length=80, description="Pricing trend information")

    @validator('sale_price')
    def validate_sale_price(cls, v, values):
        if v and values.get('price') and v.value > values['price'].value:
            raise ValueError('Sale price cannot be greater than regular price')
        return v

    @validator('sale_price_effective_date')
    def validate_sale_date_required_with_sale_price(cls, v, values):
        if values.get('sale_price') and not v:
            raise ValueError('Sale price effective date is required when sale price is provided')
        return v


class AvailabilityInventory(BaseModel):
    """Current stock levels and timing signals"""
    availability: AvailabilityStatus = Field(..., description="Product availability")
    availability_date: Optional[date] = Field(None, description="Availability date if preorder")
    inventory_quantity: int = Field(..., ge=0, description="Stock count")
    expiration_date: Optional[date] = Field(None, description="Remove product after date")
    pickup_method: Optional[PickupMethod] = Field(None, description="Pickup options")
    pickup_sla: Optional[str] = Field(None, description="Pickup SLA")

    @validator('availability_date')
    def validate_availability_date_required_for_preorder(cls, v, values):
        if values.get('availability') == AvailabilityStatus.PREORDER and not v:
            raise ValueError('Availability date is required for preorder items')
        return v


class Variants(BaseModel):
    """Variant relationships and distinguishing attributes"""
    item_group_id: Optional[str] = Field(None, max_length=70, description="Variant group ID")
    item_group_title: Optional[str] = Field(None, max_length=150, description="Group product title")
    color: Optional[str] = Field(None, max_length=40, description="Variant color")
    size: Optional[str] = Field(None, max_length=20, description="Variant size")
    size_system: Optional[str] = Field(None, min_length=2, max_length=2, description="Size system country code")
    gender: Optional[Gender] = Field(None, description="Gender target")
    offer_id: Optional[str] = Field(None, description="Offer ID (SKU+seller+price)")
    custom_variant1_category: Optional[str] = Field(None, description="Custom variant dimension 1")
    custom_variant1_option: Optional[str] = Field(None, description="Custom variant 1 option")
    custom_variant2_category: Optional[str] = Field(None, description="Custom variant dimension 2")
    custom_variant2_option: Optional[str] = Field(None, description="Custom variant 2 option")
    custom_variant3_category: Optional[str] = Field(None, description="Custom variant dimension 3")
    custom_variant3_option: Optional[str] = Field(None, description="Custom variant 3 option")


class Fulfillment(BaseModel):
    """Shipping methods and delivery times"""
    shipping: Optional[List[str]] = Field(None, description="Shipping method/cost/region")
    delivery_estimate: Optional[date] = Field(None, description="Estimated arrival date")


class MerchantInfo(BaseModel):
    """Seller identification and policies"""
    seller_name: str = Field(..., max_length=70, description="Seller name")
    seller_url: HttpUrl = Field(..., description="Seller page URL")
    seller_privacy_policy: Optional[HttpUrl] = Field(None, description="Seller-specific privacy policy")
    seller_tos: Optional[HttpUrl] = Field(None, description="Seller-specific terms of service")


class Returns(BaseModel):
    """Return policies and time windows"""
    return_policy: HttpUrl = Field(..., description="Return policy URL")
    return_window: int = Field(..., gt=0, description="Days allowed for return")


class PerformanceSignals(BaseModel):
    """Popularity and return-rate metrics"""
    popularity_score: Optional[float] = Field(None, ge=0, le=5, description="Popularity indicator (0-5 scale)")
    return_rate: Optional[float] = Field(None, ge=0, le=100, description="Return rate (0-100%)")


class Compliance(BaseModel):
    """Regulatory warnings and age restrictions"""
    warning: Optional[str] = Field(None, max_length=200, description="Product disclaimers")
    warning_url: Optional[HttpUrl] = Field(None, description="Warning URL")
    age_restriction: Optional[int] = Field(None, gt=0, description="Minimum purchase age")


class ReviewsQA(BaseModel):
    """Aggregated review statistics and FAQ content"""
    product_review_count: Optional[int] = Field(None, ge=0, description="Number of product reviews")
    product_review_rating: Optional[float] = Field(None, ge=0, le=5, description="Average review score (0-5)")
    store_review_count: Optional[int] = Field(None, ge=0, description="Number of brand/store reviews")
    store_review_rating: Optional[float] = Field(None, ge=0, le=5, description="Average store rating (0-5)")
    q_and_a: Optional[str] = Field(None, description="FAQ content")
    raw_review_data: Optional[str] = Field(None, description="Raw review payload")


class RelatedProducts(BaseModel):
    """Products commonly bought together or substitutes"""
    related_product_id: Optional[List[str]] = Field(None, description="Associated product IDs")
    relationship_type: Optional[RelationshipType] = Field(None, description="Relationship type")


class GeoTagging(BaseModel):
    """Region-specific pricing and availability"""
    geo_price: Optional[Money] = Field(None, description="Region-specific price")
    geo_availability: Optional[str] = Field(None, description="Region-specific availability")


class ProductFeed(BaseModel):
    """Complete product feed entry according to OpenAI Product Feed Specification"""
    openai_flags: OpenAIFlags = Field(..., description="OpenAI-specific flags")
    basic_data: BasicProductData = Field(..., description="Basic product information")
    item_info: ItemInformation = Field(..., description="Item-specific information")
    media: Media = Field(..., description="Media assets")
    price_promotions: PricePromotions = Field(..., description="Pricing information")
    availability_inventory: AvailabilityInventory = Field(..., description="Availability and inventory")
    variants: Variants = Field(..., description="Product variants")
    fulfillment: Fulfillment = Field(..., description="Fulfillment options")
    merchant_info: MerchantInfo = Field(..., description="Merchant information")
    returns: Returns = Field(..., description="Return policies")
    performance_signals: Optional[PerformanceSignals] = Field(None, description="Performance metrics")
    compliance: Optional[Compliance] = Field(None, description="Compliance information")
    reviews_qa: Optional[ReviewsQA] = Field(None, description="Reviews and Q&A")
    related_products: Optional[RelatedProducts] = Field(None, description="Related products")
    geo_tagging: Optional[GeoTagging] = Field(None, description="Geo-specific information")


# ==================== FEED PROCESSING MODELS ====================

class FeedIngestionRequest(BaseModel):
    """Request to ingest a product feed"""
    merchant_id: str = Field(..., description="Merchant identifier")
    feed_data: List[ProductFeed] = Field(..., description="Product feed data")
    format: Literal["json", "csv", "tsv", "xml"] = Field(..., description="Feed format")
    version: Optional[str] = Field(None, description="Feed version")
    metadata: Optional[Dict[str, Any]] = Field(None, description="Additional metadata")


class FeedIngestionResponse(BaseModel):
    """Response from feed ingestion"""
    ok: bool = Field(..., description="Whether ingestion was successful")
    processed_count: int = Field(..., ge=0, description="Number of products processed")
    error_count: int = Field(..., ge=0, description="Number of errors encountered")
    errors: Optional[List[str]] = Field(None, description="List of error messages")
    feed_id: Optional[str] = Field(None, description="Generated feed ID")
    message: Optional[str] = Field(None, description="Status message")


class ValidationError(BaseModel):
    """Validation error details"""
    field: str = Field(..., description="Field name")
    message: str = Field(..., description="Error message")
    product_id: Optional[str] = Field(None, description="Product ID")
    severity: str = Field(..., description="Error severity")
    code: str = Field(..., description="Error code")


class ValidationWarning(BaseModel):
    """Validation warning details"""
    field: str = Field(..., description="Field name")
    message: str = Field(..., description="Warning message")
    product_id: Optional[str] = Field(None, description="Product ID")
    code: str = Field(..., description="Warning code")


class FeedValidationResult(BaseModel):
    """Validation results for a product feed"""
    valid: bool = Field(..., description="Whether feed is valid")
    errors: Optional[List[ValidationError]] = Field(None, description="Validation errors")
    warnings: Optional[List[ValidationWarning]] = Field(None, description="Validation warnings")
    product_count: int = Field(..., ge=0, description="Total number of products")
    valid_products: int = Field(..., ge=0, description="Number of valid products")


class FeedStatus(BaseModel):
    """Status of a product feed"""
    feed_id: str = Field(..., description="Feed identifier")
    merchant_id: str = Field(..., description="Merchant identifier")
    status: Literal["pending", "processing", "completed", "failed"] = Field(..., description="Processing status")
    processed_at: Optional[datetime] = Field(None, description="Processing completion time")
    product_count: int = Field(..., ge=0, description="Number of products")
    error_count: int = Field(..., ge=0, description="Number of errors")
    last_updated: datetime = Field(default_factory=datetime.utcnow, description="Last update time")


# ==================== SEARCH MODELS ====================

class ProductSearchRequest(BaseModel):
    """Request to search products"""
    query: Optional[str] = Field(None, description="Search query")
    category: Optional[str] = Field(None, description="Product category filter")
    brand: Optional[str] = Field(None, description="Brand filter")
    min_price: Optional[Money] = Field(None, description="Minimum price filter")
    max_price: Optional[Money] = Field(None, description="Maximum price filter")
    availability: Optional[AvailabilityStatus] = Field(None, description="Availability filter")
    merchant_id: Optional[str] = Field(None, description="Merchant filter")
    limit: Optional[int] = Field(20, ge=1, le=100, description="Number of results to return")
    offset: Optional[int] = Field(0, ge=0, description="Number of results to skip")
    sort_by: Optional[str] = Field(None, description="Sort field")
    sort_order: Optional[Literal["asc", "desc"]] = Field("asc", description="Sort order")
    filters: Optional[Dict[str, Any]] = Field(None, description="Additional filters")


class EnhancedProduct(BaseModel):
    """Enhanced product with OpenAI Product Feed Specification compliance"""
    product_feed: ProductFeed = Field(..., description="Complete product feed data")
    created_at: datetime = Field(default_factory=datetime.utcnow, description="Creation timestamp")
    updated_at: datetime = Field(default_factory=datetime.utcnow, description="Last update timestamp")
    merchant_id: str = Field(..., description="Merchant identifier")
    status: Literal["active", "inactive", "pending_review"] = Field(..., description="Product status")


class ProductSearchResponse(BaseModel):
    """Response from product search"""
    products: List[EnhancedProduct] = Field(..., description="Search results")
    total: int = Field(..., ge=0, description="Total number of results")
    limit: int = Field(..., ge=1, description="Number of results returned")
    offset: int = Field(..., ge=0, description="Number of results skipped")
    has_more: bool = Field(..., description="Whether more results are available")


# ==================== LEGACY COMPATIBILITY ====================

class ProductVariant(BaseModel):
    """Legacy product variant model"""
    id: str = Field(..., description="Variant identifier")
    size: Optional[str] = Field(None, description="Variant size")
    price: Money = Field(..., description="Variant price")


class Product(BaseModel):
    """Legacy product model for backward compatibility"""
    id: str = Field(..., description="Product identifier")
    title: str = Field(..., description="Product title")
    description: Optional[str] = Field(None, description="Product description")
    variants: List[ProductVariant] = Field(default_factory=list, description="Product variants")
    price: Money = Field(..., description="Product price")
    available: bool = Field(..., description="Product availability")
    images: Optional[List[str]] = Field(None, description="Product images")
    tags: Optional[List[str]] = Field(None, description="Product tags")


# ==================== EXAMPLE DATA ====================

def example_product_feed() -> ProductFeed:
    """Create an example product feed for testing"""
    return ProductFeed(
        openai_flags=OpenAIFlags(
            enable_search=True,
            enable_checkout=True
        ),
        basic_data=BasicProductData(
            id="SKU12345",
            gtin="1234567890123",
            mpn="GPT5",
            title="Men's Trail Running Shoes",
            description="Waterproof trail shoe with cushioned sole designed for rugged terrain and long-distance running.",
            link="https://example.com/product/SKU12345"
        ),
        item_info=ItemInformation(
            condition=ProductCondition.NEW,
            product_category="Apparel & Accessories > Shoes",
            brand="OpenAI",
            material="Synthetic Leather",
            dimensions="12x8x5 in",
            weight="1.5 lb",
            age_group=AgeGroup.ADULT
        ),
        media=Media(
            image_link="https://example.com/images/shoe-main.jpg",
            additional_image_link=[
                "https://example.com/images/shoe-side.jpg",
                "https://example.com/images/shoe-back.jpg"
            ],
            video_link="https://youtu.be/12345",
            model_3d_link="https://example.com/models/shoe.glb"
        ),
        price_promotions=PricePromotions(
            price=Money(value=79.99, currency="USD"),
            sale_price=Money(value=59.99, currency="USD"),
            sale_price_effective_date="2025-07-01 / 2025-07-15",
            unit_pricing_measure="1 pair",
            base_measure="1 pair",
            pricing_trend="Lowest price in 6 months"
        ),
        availability_inventory=AvailabilityInventory(
            availability=AvailabilityStatus.IN_STOCK,
            inventory_quantity=25,
            pickup_method=PickupMethod.IN_STORE,
            pickup_sla="1 day"
        ),
        variants=Variants(
            item_group_id="SHOE123GROUP",
            item_group_title="Men's Trail Running Shoes",
            color="Black",
            size="10",
            size_system="US",
            gender=Gender.MALE,
            offer_id="SKU12345-Black-79.99"
        ),
        fulfillment=Fulfillment(
            shipping=[
                "US:CA:Overnight:16.00 USD",
                "US:CA:Standard:8.00 USD",
                "US:CA:Economy:4.00 USD"
            ],
            delivery_estimate=date(2025, 8, 12)
        ),
        merchant_info=MerchantInfo(
            seller_name="Example Store",
            seller_url="https://example.com/store",
            seller_privacy_policy="https://example.com/privacy",
            seller_tos="https://example.com/terms"
        ),
        returns=Returns(
            return_policy="https://example.com/returns",
            return_window=30
        ),
        performance_signals=PerformanceSignals(
            popularity_score=4.7,
            return_rate=2.0
        ),
        compliance=Compliance(
            warning="Contains lithium battery",
            warning_url="https://example.com/warnings/battery"
        ),
        reviews_qa=ReviewsQA(
            product_review_count=254,
            product_review_rating=4.6,
            store_review_count=2000,
            store_review_rating=4.8,
            q_and_a="Q: Is this waterproof? A: Yes, these shoes are fully waterproof."
        ),
        related_products=RelatedProducts(
            related_product_id=["SKU67890", "SKU11111"],
            relationship_type=RelationshipType.OFTEN_BOUGHT_WITH
        ),
        geo_tagging=GeoTagging(
            geo_price=Money(value=79.99, currency="USD"),
            geo_availability="in_stock (California), out_of_stock (New York)"
        )
    )

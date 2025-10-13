"""
Product Feed Validation and Processing for FastAPI

This module provides validation logic and feed processing capabilities for the OpenAI Product Feed Specification.
"""

import csv
import json
import xml.etree.ElementTree as ET
from typing import List, Dict, Any, Optional, Tuple
from datetime import datetime, date
import io
import re

try:
    # Try relative imports first (when running as module)
    from .product_feed_models import (
        ProductFeed, FeedIngestionRequest, FeedIngestionResponse, 
        FeedValidationResult, ValidationError, ValidationWarning,
        FeedStatus, EnhancedProduct, Money
    )
except ImportError:
    # Fall back to absolute imports (when running directly)
    from product_feed_models import (
        ProductFeed, FeedIngestionRequest, FeedIngestionResponse, 
        FeedValidationResult, ValidationError, ValidationWarning,
        FeedStatus, EnhancedProduct, Money
    )


class ProductFeedValidator:
    """Validates product feed data according to OpenAI Product Feed Specification"""
    
    def __init__(self):
        self.valid_currencies = {
            'USD', 'EUR', 'GBP', 'CAD', 'AUD', 'JPY', 'CHF', 'CNY', 'INR', 'BRL', 'MXN',
            'KRW', 'SGD', 'HKD', 'NOK', 'SEK', 'DKK', 'PLN', 'CZK', 'HUF', 'ILS', 'CLP',
            'PHP', 'AED', 'SAR', 'ZAR', 'THB', 'MYR', 'IDR', 'VND', 'TRY', 'RUB', 'UAH',
            'KZT', 'BGN', 'RON', 'HRK', 'ISK', 'NZD', 'EGP', 'QAR', 'KWD', 'BHD', 'OMR',
            'JOD', 'LBP', 'PKR', 'BDT', 'LKR', 'NPR', 'AFN', 'AMD', 'AZN', 'GEL', 'KGS',
            'TJS', 'TMT', 'UZS'
        }
    
    def validate_product_feed(self, feed: ProductFeed) -> FeedValidationResult:
        """Validate a complete product feed entry"""
        errors = []
        warnings = []
        
        # Validate OpenAI flags
        self._validate_openai_flags(feed.openai_flags, feed.basic_data.id, errors)
        
        # Validate basic product data
        self._validate_basic_data(feed.basic_data, errors, warnings)
        
        # Validate item information
        self._validate_item_info(feed.item_info, feed.basic_data.id, errors, warnings)
        
        # Validate media
        self._validate_media(feed.media, feed.basic_data.id, errors)
        
        # Validate pricing
        self._validate_pricing(feed.price_promotions, feed.basic_data.id, errors)
        
        # Validate availability
        self._validate_availability(feed.availability_inventory, feed.basic_data.id, errors)
        
        # Validate variants
        self._validate_variants(feed.variants, feed.basic_data.id, errors)
        
        # Validate fulfillment
        self._validate_fulfillment(feed.fulfillment, feed.basic_data.id, errors)
        
        # Validate merchant info
        self._validate_merchant_info(feed.merchant_info, feed.basic_data.id, errors)
        
        # Validate returns
        self._validate_returns(feed.returns, feed.basic_data.id, errors)
        
        # Validate optional fields if present
        if feed.performance_signals:
            self._validate_performance_signals(feed.performance_signals, feed.basic_data.id, errors)
        
        if feed.compliance:
            self._validate_compliance(feed.compliance, feed.basic_data.id, errors)
        
        if feed.reviews_qa:
            self._validate_reviews_qa(feed.reviews_qa, feed.basic_data.id, errors)
        
        if feed.related_products:
            self._validate_related_products(feed.related_products, feed.basic_data.id, errors)
        
        if feed.geo_tagging:
            self._validate_geo_tagging(feed.geo_tagging, feed.basic_data.id, errors)
        
        return FeedValidationResult(
            valid=len(errors) == 0,
            errors=errors if errors else None,
            warnings=warnings if warnings else None,
            product_count=1,
            valid_products=1 if len(errors) == 0 else 0
        )
    
    def validate_feed_ingestion_request(self, request: FeedIngestionRequest) -> FeedValidationResult:
        """Validate a feed ingestion request"""
        errors = []
        warnings = []
        
        if not request.merchant_id:
            errors.append(ValidationError(
                field="merchant_id",
                message="Merchant ID is required",
                severity="error",
                code="REQUIRED_FIELD"
            ))
        
        if not request.feed_data:
            errors.append(ValidationError(
                field="feed_data",
                message="Feed data is required",
                severity="error",
                code="REQUIRED_FIELD"
            ))
        
        valid_products = 0
        for i, product in enumerate(request.feed_data):
            result = self.validate_product_feed(product)
            if result.errors:
                for error in result.errors:
                    error.product_id = product.basic_data.id
                    errors.append(error)
            if result.warnings:
                for warning in result.warnings:
                    warning.product_id = product.basic_data.id
                    warnings.append(warning)
            if result.valid:
                valid_products += 1
        
        return FeedValidationResult(
            valid=len(errors) == 0,
            errors=errors if errors else None,
            warnings=warnings if warnings else None,
            product_count=len(request.feed_data),
            valid_products=valid_products
        )
    
    def _validate_openai_flags(self, flags, product_id: str, errors: List[ValidationError]):
        """Validate OpenAI flags"""
        if not flags.enable_search and flags.enable_checkout:
            errors.append(ValidationError(
                field="enable_checkout",
                message="enable_checkout requires enable_search to be true",
                product_id=product_id,
                severity="error",
                code="INVALID_VALUE"
            ))
    
    def _validate_basic_data(self, data, errors: List[ValidationError], warnings: List[ValidationWarning]):
        """Validate basic product data"""
        if not data.id:
            errors.append(ValidationError(
                field="id",
                message="Product ID is required",
                severity="error",
                code="REQUIRED_FIELD"
            ))
        
        if not data.title:
            errors.append(ValidationError(
                field="title",
                message="Title is required",
                severity="error",
                code="REQUIRED_FIELD"
            ))
        
        if not data.description:
            errors.append(ValidationError(
                field="description",
                message="Description is required",
                severity="error",
                code="REQUIRED_FIELD"
            ))
        
        if not data.link:
            errors.append(ValidationError(
                field="link",
                message="Product link is required",
                severity="error",
                code="REQUIRED_FIELD"
            ))
        
        # Validate GTIN format if provided
        if data.gtin and not re.match(r'^\d{8}$|^\d{12}$|^\d{13}$|^\d{14}$', data.gtin):
            errors.append(ValidationError(
                field="gtin",
                message="Invalid GTIN format (must be 8, 12, 13, or 14 digits)",
                severity="error",
                code="INVALID_GTIN"
            ))
        
        # Warn if GTIN is missing (recommended field)
        if not data.gtin:
            warnings.append(ValidationWarning(
                field="gtin",
                message="GTIN is recommended for better product identification",
                code="MISSING_RECOMMENDED_FIELD"
            ))
    
    def _validate_item_info(self, info, product_id: str, errors: List[ValidationError], warnings: List[ValidationWarning]):
        """Validate item information"""
        if not info.product_category:
            errors.append(ValidationError(
                field="product_category",
                message="Product category is required",
                product_id=product_id,
                severity="error",
                code="REQUIRED_FIELD"
            ))
        
        # Warn if brand is missing (recommended field)
        if not info.brand:
            warnings.append(ValidationWarning(
                field="brand",
                message="Brand is recommended for better product discoverability",
                product_id=product_id,
                code="MISSING_RECOMMENDED_FIELD"
            ))
    
    def _validate_media(self, media, product_id: str, errors: List[ValidationError]):
        """Validate media assets"""
        if not media.image_link:
            errors.append(ValidationError(
                field="image_link",
                message="Main image link is required",
                product_id=product_id,
                severity="error",
                code="REQUIRED_FIELD"
            ))
    
    def _validate_pricing(self, pricing, product_id: str, errors: List[ValidationError]):
        """Validate pricing information"""
        if not pricing.price:
            errors.append(ValidationError(
                field="price",
                message="Price is required",
                product_id=product_id,
                severity="error",
                code="REQUIRED_FIELD"
            ))
        elif pricing.price.value <= 0:
            errors.append(ValidationError(
                field="price.value",
                message="Price value must be greater than 0",
                product_id=product_id,
                severity="error",
                code="INVALID_VALUE"
            ))
        
        if pricing.sale_price and pricing.price:
            if pricing.sale_price.value > pricing.price.value:
                errors.append(ValidationError(
                    field="sale_price.value",
                    message="Sale price cannot be greater than regular price",
                    product_id=product_id,
                    severity="error",
                    code="INVALID_VALUE"
                ))
            
            if pricing.sale_price.currency != pricing.price.currency:
                errors.append(ValidationError(
                    field="sale_price.currency",
                    message="Sale price currency must match regular price currency",
                    product_id=product_id,
                    severity="error",
                    code="INVALID_VALUE"
                ))
    
    def _validate_availability(self, availability, product_id: str, errors: List[ValidationError]):
        """Validate availability and inventory"""
        if availability.inventory_quantity < 0:
            errors.append(ValidationError(
                field="inventory_quantity",
                message="Inventory quantity must be non-negative",
                product_id=product_id,
                severity="error",
                code="INVALID_VALUE"
            ))
    
    def _validate_variants(self, variants, product_id: str, errors: List[ValidationError]):
        """Validate variant information"""
        # Additional variant validation can be added here
        pass
    
    def _validate_fulfillment(self, fulfillment, product_id: str, errors: List[ValidationError]):
        """Validate fulfillment information"""
        # Additional fulfillment validation can be added here
        pass
    
    def _validate_merchant_info(self, merchant, product_id: str, errors: List[ValidationError]):
        """Validate merchant information"""
        if not merchant.seller_name:
            errors.append(ValidationError(
                field="seller_name",
                message="Seller name is required",
                product_id=product_id,
                severity="error",
                code="REQUIRED_FIELD"
            ))
        
        if not merchant.seller_url:
            errors.append(ValidationError(
                field="seller_url",
                message="Seller URL is required",
                product_id=product_id,
                severity="error",
                code="REQUIRED_FIELD"
            ))
    
    def _validate_returns(self, returns, product_id: str, errors: List[ValidationError]):
        """Validate return policy information"""
        if not returns.return_policy:
            errors.append(ValidationError(
                field="return_policy",
                message="Return policy URL is required",
                product_id=product_id,
                severity="error",
                code="REQUIRED_FIELD"
            ))
        
        if returns.return_window <= 0:
            errors.append(ValidationError(
                field="return_window",
                message="Return window must be a positive integer",
                product_id=product_id,
                severity="error",
                code="INVALID_VALUE"
            ))
    
    def _validate_performance_signals(self, signals, product_id: str, errors: List[ValidationError]):
        """Validate performance signals"""
        if signals.popularity_score is not None and (signals.popularity_score < 0 or signals.popularity_score > 5):
            errors.append(ValidationError(
                field="popularity_score",
                message="Popularity score must be between 0 and 5",
                product_id=product_id,
                severity="error",
                code="INVALID_RATING"
            ))
        
        if signals.return_rate is not None and (signals.return_rate < 0 or signals.return_rate > 100):
            errors.append(ValidationError(
                field="return_rate",
                message="Return rate must be between 0 and 100",
                product_id=product_id,
                severity="error",
                code="INVALID_PERCENTAGE"
            ))
    
    def _validate_compliance(self, compliance, product_id: str, errors: List[ValidationError]):
        """Validate compliance information"""
        if compliance.age_restriction is not None and compliance.age_restriction < 0:
            errors.append(ValidationError(
                field="age_restriction",
                message="Age restriction must be a positive integer",
                product_id=product_id,
                severity="error",
                code="INVALID_VALUE"
            ))
    
    def _validate_reviews_qa(self, reviews, product_id: str, errors: List[ValidationError]):
        """Validate reviews and Q&A"""
        if reviews.product_review_count is not None and reviews.product_review_count < 0:
            errors.append(ValidationError(
                field="product_review_count",
                message="Product review count must be non-negative",
                product_id=product_id,
                severity="error",
                code="INVALID_VALUE"
            ))
        
        if reviews.product_review_rating is not None and (reviews.product_review_rating < 0 or reviews.product_review_rating > 5):
            errors.append(ValidationError(
                field="product_review_rating",
                message="Product review rating must be between 0 and 5",
                product_id=product_id,
                severity="error",
                code="INVALID_RATING"
            ))
    
    def _validate_related_products(self, related, product_id: str, errors: List[ValidationError]):
        """Validate related products"""
        # Additional related products validation can be added here
        pass
    
    def _validate_geo_tagging(self, geo, product_id: str, errors: List[ValidationError]):
        """Validate geo tagging"""
        if geo.geo_price and geo.geo_price.value <= 0:
            errors.append(ValidationError(
                field="geo_price.value",
                message="Geo price value must be greater than 0",
                product_id=product_id,
                severity="error",
                code="INVALID_VALUE"
            ))


class ProductFeedProcessor:
    """Processes product feeds in different formats"""
    
    def __init__(self):
        self.validator = ProductFeedValidator()
    
    def process_feed(self, format_type: str, data: bytes) -> FeedIngestionRequest:
        """Process a product feed in the specified format"""
        if format_type.lower() == "json":
            return self._process_json_feed(data)
        elif format_type.lower() == "csv":
            return self._process_csv_feed(data)
        elif format_type.lower() == "tsv":
            return self._process_tsv_feed(data)
        elif format_type.lower() == "xml":
            return self._process_xml_feed(data)
        else:
            raise ValueError(f"Unsupported format: {format_type}")
    
    def _process_json_feed(self, data: bytes) -> FeedIngestionRequest:
        """Process JSON formatted feed"""
        try:
            feed_data = json.loads(data.decode('utf-8'))
            if isinstance(feed_data, dict):
                # Single product
                feed_data = [feed_data]
            elif isinstance(feed_data, list):
                # Multiple products
                pass
            else:
                raise ValueError("Invalid JSON format")
            
            products = []
            for item in feed_data:
                # Convert dict to ProductFeed model
                product = ProductFeed(**item)
                products.append(product)
            
            return FeedIngestionRequest(
                merchant_id="json_merchant",  # This should be provided in the request context
                feed_data=products,
                format="json",
                version="1.0",
                metadata={
                    "source": "json_upload",
                    "processed_at": datetime.utcnow().isoformat()
                }
            )
        except Exception as e:
            raise ValueError(f"Failed to process JSON feed: {str(e)}")
    
    def _process_csv_feed(self, data: bytes) -> FeedIngestionRequest:
        """Process CSV formatted feed"""
        try:
            content = data.decode('utf-8')
            reader = csv.DictReader(io.StringIO(content))
            
            products = []
            for row in reader:
                product = self._csv_row_to_product_feed(row)
                products.append(product)
            
            return FeedIngestionRequest(
                merchant_id="csv_merchant",  # This should be provided in the request context
                feed_data=products,
                format="csv",
                version="1.0",
                metadata={
                    "source": "csv_upload",
                    "processed_at": datetime.utcnow().isoformat()
                }
            )
        except Exception as e:
            raise ValueError(f"Failed to process CSV feed: {str(e)}")
    
    def _process_tsv_feed(self, data: bytes) -> FeedIngestionRequest:
        """Process TSV formatted feed"""
        try:
            content = data.decode('utf-8')
            reader = csv.DictReader(io.StringIO(content), delimiter='\t')
            
            products = []
            for row in reader:
                product = self._csv_row_to_product_feed(row)
                products.append(product)
            
            return FeedIngestionRequest(
                merchant_id="tsv_merchant",  # This should be provided in the request context
                feed_data=products,
                format="tsv",
                version="1.0",
                metadata={
                    "source": "tsv_upload",
                    "processed_at": datetime.utcnow().isoformat()
                }
            )
        except Exception as e:
            raise ValueError(f"Failed to process TSV feed: {str(e)}")
    
    def _process_xml_feed(self, data: bytes) -> FeedIngestionRequest:
        """Process XML formatted feed"""
        try:
            root = ET.fromstring(data)
            
            products = []
            for product_elem in root.findall('.//product'):
                product_dict = self._xml_element_to_dict(product_elem)
                product = ProductFeed(**product_dict)
                products.append(product)
            
            merchant_id = root.get('merchant_id', 'xml_merchant')
            
            return FeedIngestionRequest(
                merchant_id=merchant_id,
                feed_data=products,
                format="xml",
                version="1.0",
                metadata={
                    "source": "xml_upload",
                    "processed_at": datetime.utcnow().isoformat()
                }
            )
        except Exception as e:
            raise ValueError(f"Failed to process XML feed: {str(e)}")
    
    def _csv_row_to_product_feed(self, row: Dict[str, str]) -> ProductFeed:
        """Convert CSV row to ProductFeed"""
        # This is a simplified conversion - in production, you'd want more robust mapping
        return ProductFeed(
            openai_flags={
                "enable_search": self._parse_bool(row.get("enable_search", "true")),
                "enable_checkout": self._parse_bool(row.get("enable_checkout", "false"))
            },
            basic_data={
                "id": row.get("id", ""),
                "gtin": row.get("gtin"),
                "mpn": row.get("mpn"),
                "title": row.get("title", ""),
                "description": row.get("description", ""),
                "link": row.get("link", "")
            },
            item_info={
                "condition": row.get("condition"),
                "product_category": row.get("product_category", ""),
                "brand": row.get("brand"),
                "material": row.get("material"),
                "weight": row.get("weight")
            },
            media={
                "image_link": row.get("image_link", ""),
                "additional_image_link": self._parse_list(row.get("additional_image_link"))
            },
            price_promotions={
                "price": {
                    "value": float(row.get("price_value", 0)),
                    "currency": row.get("price_currency", "USD")
                },
                "sale_price": self._parse_money(row.get("sale_price_value"), row.get("sale_price_currency")),
                "sale_price_effective_date": row.get("sale_price_effective_date")
            },
            availability_inventory={
                "availability": row.get("availability", "in_stock"),
                "inventory_quantity": int(row.get("inventory_quantity", 0))
            },
            variants={
                "color": row.get("color"),
                "size": row.get("size"),
                "gender": row.get("gender")
            },
            fulfillment={
                "shipping": self._parse_list(row.get("shipping"))
            },
            merchant_info={
                "seller_name": row.get("seller_name", ""),
                "seller_url": row.get("seller_url", "")
            },
            returns={
                "return_policy": row.get("return_policy", ""),
                "return_window": int(row.get("return_window", 30))
            }
        )
    
    def _xml_element_to_dict(self, element) -> Dict[str, Any]:
        """Convert XML element to dictionary"""
        # Simplified XML to dict conversion
        result = {}
        for child in element:
            if len(child) == 0:
                result[child.tag] = child.text
            else:
                result[child.tag] = self._xml_element_to_dict(child)
        return result
    
    def _parse_bool(self, value: str) -> bool:
        """Parse boolean value from string"""
        if not value:
            return False
        return value.lower() in ['true', '1', 'yes', 'y']
    
    def _parse_list(self, value: str) -> Optional[List[str]]:
        """Parse comma-separated list from string"""
        if not value:
            return None
        return [item.strip() for item in value.split(',') if item.strip()]
    
    def _parse_money(self, value_str: str, currency_str: str) -> Optional[Dict[str, Any]]:
        """Parse money value from strings"""
        if not value_str or not currency_str:
            return None
        try:
            return {
                "value": float(value_str),
                "currency": currency_str
            }
        except ValueError:
            return None


class ProductFeedTransformer:
    """Transforms between different product representations"""
    
    @staticmethod
    def product_feed_to_enhanced_product(feed: ProductFeed, merchant_id: str) -> EnhancedProduct:
        """Convert ProductFeed to EnhancedProduct"""
        return EnhancedProduct(
            product_feed=feed,
            merchant_id=merchant_id,
            status="active"
        )
    
    @staticmethod
    def product_feed_to_legacy_product(feed: ProductFeed) -> Dict[str, Any]:
        """Convert ProductFeed to legacy Product format"""
        # Extract main image
        images = [str(feed.media.image_link)]
        if feed.media.additional_image_link:
            images.extend([str(img) for img in feed.media.additional_image_link])
        
        # Create variants from feed data
        variants = []
        if feed.variants.item_group_id:
            variants.append({
                "id": feed.basic_data.id,
                "size": feed.variants.size,
                "price": {
                    "value": feed.price_promotions.price.value,
                    "currency": feed.price_promotions.price.currency
                }
            })
        
        return {
            "id": feed.basic_data.id,
            "title": feed.basic_data.title,
            "description": feed.basic_data.description,
            "variants": variants,
            "price": {
                "value": feed.price_promotions.price.value,
                "currency": feed.price_promotions.price.currency
            },
            "available": feed.availability_inventory.availability == "in_stock",
            "images": images,
            "tags": [feed.item_info.product_category, feed.item_info.brand] if feed.item_info.brand else [feed.item_info.product_category]
        }


class FeedStorage:
    """Manages feed storage and status tracking"""
    
    def __init__(self):
        self.feeds: Dict[str, FeedStatus] = {}
    
    def store_feed(self, feed_id: str, status: FeedStatus):
        """Store a feed status"""
        self.feeds[feed_id] = status
    
    def get_feed(self, feed_id: str) -> Optional[FeedStatus]:
        """Retrieve a feed status"""
        return self.feeds.get(feed_id)
    
    def update_feed_status(self, feed_id: str, status: str, product_count: int, error_count: int):
        """Update feed status"""
        if feed_id in self.feeds:
            feed = self.feeds[feed_id]
            feed.status = status
            feed.product_count = product_count
            feed.error_count = error_count
            feed.last_updated = datetime.utcnow()
            if status in ["completed", "failed"]:
                feed.processed_at = datetime.utcnow()
    
    def list_feeds(self) -> List[FeedStatus]:
        """List all feed statuses"""
        return list(self.feeds.values())

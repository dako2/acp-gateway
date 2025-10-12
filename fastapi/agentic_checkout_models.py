"""
Agentic Checkout Specification Models for FastAPI
Implements the complete ACP Agentic Checkout API
"""

from pydantic import BaseModel, Field, validator, HttpUrl
from typing import List, Optional, Union, Literal
from datetime import datetime
from enum import Enum


# ==================== ENUMS ====================

class CheckoutSessionStatus(str, Enum):
    NOT_READY_FOR_PAYMENT = "not_ready_for_payment"
    READY_FOR_PAYMENT = "ready_for_payment"
    COMPLETED = "completed"
    CANCELED = "canceled"
    IN_PROGRESS = "in_progress"


class OrderStatus(str, Enum):
    CREATED = "created"
    MANUAL_REVIEW = "manual_review"
    CONFIRMED = "confirmed"
    CANCELED = "canceled"
    SHIPPED = "shipped"
    FULFILLED = "fulfilled"


class PaymentProvider(str, Enum):
    STRIPE = "stripe"


class PaymentMethod(str, Enum):
    CARD = "card"


class TotalType(str, Enum):
    ITEMS_BASE_AMOUNT = "items_base_amount"
    ITEMS_DISCOUNT = "items_discount"
    SUBTOTAL = "subtotal"
    DISCOUNT = "discount"
    FULFILLMENT = "fulfillment"
    TAX = "tax"
    FEE = "fee"
    TOTAL = "total"


class FulfillmentType(str, Enum):
    SHIPPING = "shipping"
    DIGITAL = "digital"


class MessageType(str, Enum):
    INFO = "info"
    ERROR = "error"


class ErrorCode(str, Enum):
    MISSING = "missing"
    INVALID = "invalid"
    OUT_OF_STOCK = "out_of_stock"
    PAYMENT_DECLINED = "payment_declined"
    REQUIRES_SIGN_IN = "requires_sign_in"
    REQUIRES_3DS = "requires_3ds"


class ContentType(str, Enum):
    PLAIN = "plain"
    MARKDOWN = "markdown"


class LinkType(str, Enum):
    TERMS_OF_USE = "terms_of_use"
    PRIVACY_POLICY = "privacy_policy"
    SELLER_SHOP_POLICIES = "seller_shop_policies"


class RefundType(str, Enum):
    STORE_CREDIT = "store_credit"
    ORIGINAL_PAYMENT = "original_payment"


# ==================== CORE MODELS ====================

class Address(BaseModel):
    """Address for shipping or billing"""
    name: str = Field(..., max_length=256)
    line_one: str = Field(..., max_length=60)
    line_two: Optional[str] = Field(None, max_length=60)
    city: str = Field(..., max_length=60)
    state: str = Field(..., description="State/county/province/region")
    country: str = Field(..., min_length=2, max_length=2, description="ISO-3166-1 alpha-2")
    postal_code: str = Field(..., max_length=20)


class Buyer(BaseModel):
    """Buyer information"""
    first_name: str = Field(..., max_length=256)
    last_name: str = Field(..., max_length=256)
    email: str = Field(..., max_length=256)
    phone_number: Optional[str] = Field(None, description="E.164 format")


class Item(BaseModel):
    """Item to be purchased"""
    id: str = Field(..., description="Id of merchandise that can be purchased")
    quantity: int = Field(..., ge=1, description="Quantity of the item for fulfillment")


class PaymentProvider(BaseModel):
    """Payment provider information"""
    provider: PaymentProvider = Field(...)
    supported_payment_methods: List[PaymentMethod] = Field(...)


class LineItem(BaseModel):
    """Line item with calculated costs"""
    id: str = Field(..., description="Id of the line item (different from item id)")
    item: Item = Field(...)
    base_amount: int = Field(..., ge=0, description="Item base amount before adjustments")
    discount: int = Field(..., ge=0, description="Discount applied to the item")
    subtotal: int = Field(..., ge=0, description="Amount after all adjustments")
    tax: int = Field(..., ge=0, description="Tax amount")
    total: int = Field(..., ge=0, description="Total amount")


class Total(BaseModel):
    """Total line item"""
    type: TotalType = Field(...)
    display_text: str = Field(..., description="Text displayed to customer")
    amount: int = Field(..., ge=0, description="Amount in minor units")


class FulfillmentOptionShipping(BaseModel):
    """Shipping fulfillment option"""
    type: Literal["shipping"] = "shipping"
    id: str = Field(..., description="Unique ID for shipping option")
    title: str = Field(..., description="Title of shipping option")
    subtitle: str = Field(..., description="Estimated timeline for shipping")
    carrier: str = Field(..., description="Name of shipping carrier")
    earliest_delivery_time: datetime = Field(..., description="Earliest delivery time (RFC 3339)")
    latest_delivery_time: datetime = Field(..., description="Latest delivery time (RFC 3339)")
    subtotal: int = Field(..., ge=0, description="Subtotal cost of shipping")
    tax: int = Field(..., ge=0, description="Tax amount")
    total: int = Field(..., ge=0, description="Total cost of shipping")


class FulfillmentOptionDigital(BaseModel):
    """Digital fulfillment option"""
    type: Literal["digital"] = "digital"
    id: str = Field(..., description="Unique ID for digital option")
    title: str = Field(..., description="Title of digital option")
    subtitle: Optional[str] = Field(None, description="How item will be digitally delivered")
    subtotal: int = Field(..., ge=0, description="Subtotal cost of digital delivery")
    tax: int = Field(..., ge=0, description="Tax amount")
    total: int = Field(..., ge=0, description="Total cost of digital delivery")


FulfillmentOption = Union[FulfillmentOptionShipping, FulfillmentOptionDigital]


class MessageInfo(BaseModel):
    """Informational message"""
    type: Literal["info"] = "info"
    param: str = Field(..., description="RFC 9535 JSONPath")
    content_type: ContentType = Field(...)
    content: str = Field(..., description="Raw message content")


class MessageError(BaseModel):
    """Error message"""
    type: Literal["error"] = "error"
    code: ErrorCode = Field(...)
    param: Optional[str] = Field(None, description="RFC 9535 JSONPath")
    content_type: ContentType = Field(...)
    content: str = Field(..., description="Raw message content")


Message = Union[MessageInfo, MessageError]


class Link(BaseModel):
    """Link to policies or terms"""
    type: LinkType = Field(...)
    url: HttpUrl = Field(...)


class PaymentData(BaseModel):
    """Payment method data"""
    token: str = Field(..., description="Token representing payment method")
    provider: PaymentProvider = Field(...)
    billing_address: Optional[Address] = None


class Order(BaseModel):
    """Order created from checkout session"""
    id: str = Field(..., description="Unique order identifier")
    checkout_session_id: str = Field(..., description="Checkout session that created this order")
    permalink_url: HttpUrl = Field(..., description="URL to view order details")


class Refund(BaseModel):
    """Refund information"""
    type: RefundType = Field(...)
    amount: int = Field(..., ge=0, description="Total amount of money refunded")


# ==================== CHECKOUT SESSION MODELS ====================

class CheckoutSessionBase(BaseModel):
    """Base checkout session model"""
    id: str = Field(..., description="Unique checkout session identifier")
    buyer: Optional[Buyer] = None
    payment_provider: Optional[PaymentProvider] = None
    status: CheckoutSessionStatus = Field(...)
    currency: str = Field(..., description="ISO 4217 currency code (lowercase)")
    line_items: List[LineItem] = Field(...)
    fulfillment_address: Optional[Address] = None
    fulfillment_options: List[FulfillmentOption] = Field(...)
    fulfillment_option_id: Optional[str] = Field(None, description="Selected fulfillment option ID")
    totals: List[Total] = Field(...)
    messages: List[Message] = Field(...)
    links: List[Link] = Field(...)


class CheckoutSession(CheckoutSessionBase):
    """Standard checkout session"""
    pass


class CheckoutSessionWithOrder(CheckoutSessionBase):
    """Checkout session with completed order"""
    order: Order = Field(...)


# ==================== REQUEST MODELS ====================

class CheckoutSessionCreateRequest(BaseModel):
    """Request to create a checkout session"""
    buyer: Optional[Buyer] = None
    items: List[Item] = Field(..., min_items=1)
    fulfillment_address: Optional[Address] = None


class CheckoutSessionUpdateRequest(BaseModel):
    """Request to update a checkout session"""
    buyer: Optional[Buyer] = None
    items: Optional[List[Item]] = None
    fulfillment_address: Optional[Address] = None
    fulfillment_option_id: Optional[str] = None


class CheckoutSessionCompleteRequest(BaseModel):
    """Request to complete a checkout session"""
    buyer: Optional[Buyer] = None
    payment_data: PaymentData = Field(...)


# ==================== WEBHOOK MODELS ====================

class EventData(BaseModel):
    """Webhook event data"""
    type: Literal["order"] = "order"
    checkout_session_id: str = Field(..., description="Checkout session that created this order")
    permalink_url: HttpUrl = Field(..., description="URL that points to the order")
    status: OrderStatus = Field(..., description="Latest status of the order")
    refunds: List[Refund] = Field(..., description="List of refunds issued for the order")


class WebhookEvent(BaseModel):
    """Webhook event"""
    type: Literal["order_created", "order_updated"] = Field(..., description="Type of webhook event")
    data: EventData = Field(..., description="Webhook event data")


# ==================== ERROR MODELS ====================

class Error(BaseModel):
    """Error response"""
    type: Literal["invalid_request", "request_not_idempotent", "processing_error", "service_unavailable"] = Field(...)
    code: str = Field(..., description="Implementation-defined error code")
    message: str = Field(..., description="Human-readable error description")
    param: Optional[str] = Field(None, description="RFC 9535 JSONPath (optional)")


# ==================== EXAMPLE DATA ====================

def example_checkout_session_create_request() -> CheckoutSessionCreateRequest:
    """Create an example checkout session create request"""
    return CheckoutSessionCreateRequest(
        items=[
            Item(id="item_123", quantity=1)
        ],
        fulfillment_address=Address(
            name="John Doe",
            line_one="1234 Chat Road",
            line_two="Apt 101",
            city="San Francisco",
            state="CA",
            country="US",
            postal_code="94131"
        )
    )


def example_checkout_session() -> CheckoutSession:
    """Create an example checkout session"""
    return CheckoutSession(
        id="checkout_session_123",
        payment_provider=PaymentProvider(
            provider=PaymentProvider.STRIPE,
            supported_payment_methods=[PaymentMethod.CARD]
        ),
        status=CheckoutSessionStatus.READY_FOR_PAYMENT,
        currency="usd",
        line_items=[
            LineItem(
                id="line_item_123",
                item=Item(id="item_123", quantity=1),
                base_amount=300,
                discount=0,
                subtotal=300,
                tax=30,
                total=330
            )
        ],
        fulfillment_address=Address(
            name="John Doe",
            line_one="1234 Chat Road",
            line_two="Apt 101",
            city="San Francisco",
            state="CA",
            country="US",
            postal_code="94131"
        ),
        fulfillment_options=[
            FulfillmentOptionShipping(
                id="fulfillment_option_123",
                title="Standard",
                subtitle="Arrives in 4-5 days",
                carrier="USPS",
                earliest_delivery_time=datetime(2025, 10, 12, 7, 20, 50),
                latest_delivery_time=datetime(2025, 10, 13, 7, 20, 50),
                subtotal=100,
                tax=0,
                total=100
            )
        ],
        fulfillment_option_id="fulfillment_option_123",
        totals=[
            Total(type=TotalType.ITEMS_BASE_AMOUNT, display_text="Item(s) total", amount=300),
            Total(type=TotalType.SUBTOTAL, display_text="Subtotal", amount=300),
            Total(type=TotalType.TAX, display_text="Tax", amount=30),
            Total(type=TotalType.FULFILLMENT, display_text="Fulfillment", amount=100),
            Total(type=TotalType.TOTAL, display_text="Total", amount=430)
        ],
        messages=[],
        links=[
            Link(
                type=LinkType.TERMS_OF_USE,
                url="https://www.testshop.com/legal/terms-of-use"
            )
        ]
    )

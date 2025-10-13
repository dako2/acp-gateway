"""
Delegated Payment Specification Models for FastAPI
Implements the complete ACP Delegated Payment API for PSPs and PCI DSS Level 1 merchants
"""

from pydantic import BaseModel, Field, validator, HttpUrl
from typing import List, Optional, Union, Literal, Dict, Any
from datetime import datetime
from enum import Enum


# ==================== ENUMS ====================

class PaymentMethodType(str, Enum):
    CARD = "card"


class CardNumberType(str, Enum):
    FPAN = "fpan"
    NETWORK_TOKEN = "network_token"


class CardFundingType(str, Enum):
    CREDIT = "credit"
    DEBIT = "debit"
    PREPAID = "prepaid"


class WalletType(str, Enum):
    WALLET = "wallet"


class CardBrand(str, Enum):
    VISA = "Visa"
    AMEX = "amex"
    DISCOVER = "discover"
    MASTERCARD = "Mastercard"


class AllowanceReason(str, Enum):
    ONE_TIME = "one_time"


class RiskSignalType(str, Enum):
    CARD_TESTING = "card_testing"
    FRAUD = "fraud"
    VELOCITY = "velocity"
    GEOGRAPHIC = "geographic"
    DEVICE = "device"


class RiskAction(str, Enum):
    BLOCKED = "blocked"
    MANUAL_REVIEW = "manual_review"
    AUTHORIZED = "authorized"


class ErrorType(str, Enum):
    INVALID_REQUEST = "invalid_request"
    RATE_LIMIT_EXCEEDED = "rate_limit_exceeded"
    PROCESSING_ERROR = "processing_error"
    SERVICE_UNAVAILABLE = "service_unavailable"


class ErrorCode(str, Enum):
    INVALID_CARD = "invalid_card"
    DUPLICATE_REQUEST = "duplicate_request"
    IDEMPOTENCY_CONFLICT = "idempotency_conflict"


# ==================== CORE MODELS ====================

class PaymentMethod(BaseModel):
    """Payment method details for card transactions"""
    type: PaymentMethodType = Field(..., description="Type of payment method (currently only 'card')")
    card_number_type: CardNumberType = Field(..., description="Type of card number (FPAN or network token)")
    number: str = Field(..., description="Card number")
    exp_month: Optional[str] = Field(None, max_length=2, description="Expiry month")
    exp_year: Optional[str] = Field(None, max_length=4, description="4-digit expiry year")
    name: Optional[str] = Field(None, description="Cardholder name")
    cvc: Optional[str] = Field(None, max_length=4, description="Card CVC number")
    cryptogram: Optional[str] = Field(None, description="Cryptogram for network tokens")
    eci_value: Optional[str] = Field(None, description="Electronic Commerce Indicator for network tokens")
    checks_performed: Optional[List[str]] = Field(None, description="Checks already performed")
    iin: Optional[str] = Field(None, max_length=6, description="Institution Identification Number (BIN)")
    display_card_funding_type: CardFundingType = Field(..., description="Funding type for display")
    display_wallet_type: Optional[WalletType] = Field(None, description="Digital wallet type")
    display_brand: Optional[CardBrand] = Field(None, description="Card brand for display")
    display_last4: Optional[str] = Field(None, max_length=4, description="Last 4 digits for display")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Arbitrary key/value pairs")

    @validator('number')
    def validate_card_number(cls, v):
        if not v or len(v) < 13 or len(v) > 19:
            raise ValueError('Card number must be between 13-19 digits')
        if not v.isdigit():
            raise ValueError('Card number must contain only digits')
        return v

    @validator('exp_month')
    def validate_exp_month(cls, v):
        if v is not None:
            if not v.isdigit() or not (1 <= int(v) <= 12):
                raise ValueError('Expiry month must be between 01-12')
        return v

    @validator('exp_year')
    def validate_exp_year(cls, v):
        if v is not None:
            if not v.isdigit() or len(v) != 4:
                raise ValueError('Expiry year must be 4 digits')
            year = int(v)
            current_year = datetime.now().year
            if year < current_year or year > current_year + 10:
                raise ValueError('Expiry year must be current year or within 10 years')
        return v


class Address(BaseModel):
    """Address associated with payment method"""
    name: str = Field(..., max_length=256, description="Customer name")
    line_one: str = Field(..., max_length=60, description="Street line 1")
    line_two: Optional[str] = Field(None, max_length=60, description="Street line 2")
    city: str = Field(..., max_length=60, description="City")
    state: Optional[str] = Field(None, description="State/region (ISO-3166-2)")
    country: str = Field(..., description="ISO-3166-1 alpha-2 country code")
    postal_code: str = Field(..., max_length=20, description="Postal/ZIP code")


class Allowance(BaseModel):
    """Payment allowance constraints"""
    reason: AllowanceReason = Field(..., description="Usage reason (currently only 'one_time')")
    max_amount: int = Field(..., description="Maximum amount in minor units")
    currency: str = Field(..., description="ISO-4217 currency code")
    checkout_session_id: str = Field(..., description="Reference checkout session ID")
    merchant_id: str = Field(..., max_length=256, description="Merchant identifier")
    expires_at: str = Field(..., description="RFC 3339 expiration timestamp")

    @validator('currency')
    def validate_currency(cls, v):
        if not v or len(v) != 3:
            raise ValueError('Currency must be 3-letter ISO-4217 code')
        return v.upper()

    @validator('max_amount')
    def validate_max_amount(cls, v):
        if v <= 0:
            raise ValueError('Max amount must be positive')
        return v


class RiskSignal(BaseModel):
    """Risk assessment signal"""
    type: RiskSignalType = Field(..., description="Type of risk signal")
    score: int = Field(..., description="Risk score (0-100)")
    action: RiskAction = Field(..., description="Action taken based on risk")

    @validator('score')
    def validate_score(cls, v):
        if not (0 <= v <= 100):
            raise ValueError('Risk score must be between 0-100')
        return v


class DelegatedPaymentRequest(BaseModel):
    """Request to delegate payment processing"""
    payment_method: PaymentMethod = Field(..., description="Payment method details")
    allowance: Allowance = Field(..., description="Payment allowance constraints")
    billing_address: Optional[Address] = Field(None, description="Billing address")
    risk_signals: List[RiskSignal] = Field(..., description="Risk assessment signals")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Arbitrary key/value pairs")

    @validator('risk_signals')
    def validate_risk_signals(cls, v):
        if not v:
            raise ValueError('At least one risk signal is required')
        return v


class DelegatedPaymentResponse(BaseModel):
    """Successful delegated payment response"""
    id: str = Field(..., description="Unique vault token identifier (vt_...)")
    created: str = Field(..., description="RFC 3339 creation timestamp")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Correlation metadata")


class DelegatedPaymentError(BaseModel):
    """Error response for delegated payment"""
    type: ErrorType = Field(..., description="Error type")
    code: str = Field(..., description="Specific error code")
    message: str = Field(..., description="Human-readable error description")
    param: Optional[str] = Field(None, description="JSONPath to offending field")


# ==================== EXAMPLE DATA ====================

example_payment_method = PaymentMethod(
    type=PaymentMethodType.CARD,
    card_number_type=CardNumberType.FPAN,
    number="4242424242424242",
    exp_month="12",
    exp_year="2026",
    name="John Doe",
    cvc="123",
    display_card_funding_type=CardFundingType.CREDIT,
    display_brand=CardBrand.VISA,
    display_last4="4242",
    metadata={"issuing_bank": "test"}
)

example_allowance = Allowance(
    reason=AllowanceReason.ONE_TIME,
    max_amount=5000,  # $50.00 in cents
    currency="USD",
    checkout_session_id="checkout_session_123",
    merchant_id="merchant_abc",
    expires_at="2025-10-09T07:20:50.52Z"
)

example_billing_address = Address(
    name="John Doe",
    line_one="123 Main St",
    line_two="Apt 4B",
    city="San Francisco",
    state="CA",
    country="US",
    postal_code="94102"
)

example_risk_signal = RiskSignal(
    type=RiskSignalType.CARD_TESTING,
    score=25,
    action=RiskAction.AUTHORIZED
)

example_delegated_payment_request = DelegatedPaymentRequest(
    payment_method=example_payment_method,
    allowance=example_allowance,
    billing_address=example_billing_address,
    risk_signals=[example_risk_signal],
    metadata={"campaign": "q4", "source": "chatgpt"}
)

example_delegated_payment_response = DelegatedPaymentResponse(
    id="vt_payment_token_123",
    created="2025-01-15T10:30:00.000Z",
    metadata={"merchant_id": "merchant_abc", "source": "chatgpt"}
)

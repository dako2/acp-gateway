# ACP Gateway API Documentation

## Overview

The ACP Gateway provides a standardized interface for LLM agents to interact with e-commerce systems. It supports product browsing, cart management, checkout processing, and merchant onboarding.

## Authentication

All API endpoints (except `/healthz`) require authentication using API keys:

```
Authorization: Bearer <API_KEY>
```

API keys are configured via the `ACP_API_KEYS` environment variable (comma-separated).

## Endpoints

### Health Check

**GET** `/healthz`

Returns the health status of the gateway.

**Response:**
```json
{
  "ok": true
}
```

---

### List Products

**GET** `/acp/v1/products`

Retrieve a list of available products with optional search and pagination.

**Query Parameters:**
- `q` (optional): Search query string
- `limit` (optional, default: 20): Maximum number of products to return
- `offset` (optional, default: 0): Number of products to skip

**Response:**
```json
{
  "ok": true,
  "items": [
    {
      "id": "sku_123",
      "title": "Work Boots",
      "description": "Durable work boots for construction and outdoor work",
      "variants": [
        {
          "id": "var_9",
          "size": "9",
          "price": {
            "value": 119.99,
            "currency": "USD"
          }
        }
      ],
      "price": {
        "value": 119.99,
        "currency": "USD"
      },
      "available": true,
      "images": ["https://example.com/boots.jpg"],
      "tags": ["work", "boots", "durable"]
    }
  ],
  "total": 1
}
```

---

### Handle Intent

**POST** `/acp/v1/intent`

Process an intent from an LLM agent (add to cart, select variant, set quantity, etc.).

**Headers:**
- `X-ACP-Sandbox` (optional): Set to "true" for sandbox mode
- `Content-Type: application/json`

**Request Body:**
```json
{
  "type": "acp.intent",
  "actor": "llm",
  "payload": {
    "action": "add_to_cart",
    "items": [
      {
        "product_id": "sku_123",
        "variant_id": "var_9",
        "quantity": 2
      }
    ],
    "notes": "User prefers size 9",
    "context": {
      "session_id": "session_123",
      "locale": "en-US"
    }
  }
}
```

**Actions:**
- `add_to_cart`: Add items to the shopping cart
- `select_variant`: Update variant selection for existing items
- `set_quantity`: Update quantity for existing items
- `checkout`: Initiate checkout process (delegates to checkout endpoint)

**Response:**
```json
{
  "type": "acp.intent.result",
  "ok": true,
  "sandbox": true,
  "cart": {
    "id": "cart_001",
    "items": [
      {
        "product_id": "sku_123",
        "variant_id": "var_9",
        "quantity": 2,
        "price": {
          "value": 119.99,
          "currency": "USD"
        }
      }
    ],
    "subtotal": {
      "value": 239.98,
      "currency": "USD"
    }
  },
  "next": {
    "action": "checkout",
    "endpoint": "/acp/v1/checkout"
  }
}
```

---

### Checkout

**POST** `/acp/v1/checkout`

Create a checkout session for the specified cart.

**Headers:**
- `X-ACP-Sandbox` (optional): Set to "true" for sandbox mode
- `Content-Type: application/json`

**Request Body:**
```json
{
  "cart_id": "cart_001",
  "email": "buyer@example.com",
  "shipping_address": {
    "first_name": "John",
    "last_name": "Doe",
    "address1": "123 Main St",
    "city": "San Francisco",
    "state": "CA",
    "postal_code": "94102",
    "country": "US"
  },
  "payment": {
    "method": "card",
    "token": "tok_test_123"
  },
  "metadata": {
    "source": "llm_agent"
  }
}
```

**Response:**
```json
{
  "ok": true,
  "sandbox": true,
  "order_id": "ord_123",
  "checkout_id": "chk_123",
  "payment_intent_status": "requires_action",
  "redirect_url": "https://merchant.example/checkout/ord_123",
  "client_secret": "pi_123_secret_456",
  "total": {
    "value": 239.98,
    "currency": "USD"
  }
}
```

**Payment Intent Status:**
- `requires_action`: Customer needs to complete additional steps (3D Secure, etc.)
- `succeeded`: Payment completed successfully
- `failed`: Payment failed
- `canceled`: Payment was canceled

---

### Register Merchant

**POST** `/merchant/register`

Register a new merchant with the gateway.

**Request Body:**
```json
{
  "name": "Example Store",
  "domain": "example.com",
  "api_base": "https://example.com/api",
  "metadata": {
    "category": "retail",
    "region": "US",
    "currency": "USD"
  }
}
```

**Response:**
```json
{
  "ok": true,
  "merchant_id": "example.com"
}
```

---

### Upload Product Feed

**POST** `/merchant/feed`

Upload or refresh product data for a merchant.

**Query Parameters:**
- `url` (optional): URL to fetch product feed from

**Form Data (multipart/form-data):**
- `file` (optional): Product feed file (JSON, CSV, etc.)

**Response:**
```json
{
  "ok": true,
  "ingested": 150
}
```

---

### Webhooks

**POST** `/webhooks/acp`

Receive webhook notifications from external systems (e.g., order updates).

**Headers:**
- `X-Acp-Signature`: HMAC signature for verification
- `Content-Type: application/json`

**Request Body:**
```json
{
  "event": "order.updated",
  "data": {
    "order_id": "ord_123",
    "status": "shipped",
    "tracking_number": "1Z999AA1234567890"
  }
}
```

**Response:**
```json
{
  "ok": true
}
```

## Sandbox Mode

Sandbox mode allows testing without affecting real data or payments:

1. **Header**: `X-ACP-Sandbox: true`
2. **Query Parameter**: `?sandbox=1`

In sandbox mode:
- All transactions are marked as test/sandbox
- Payment processing is simulated
- No real charges are made
- Audit logs are clearly marked

## Error Handling

All endpoints return appropriate HTTP status codes:

- `200`: Success
- `400`: Bad Request (invalid data)
- `401`: Unauthorized (invalid API key)
- `404`: Not Found (product, cart, etc.)
- `500`: Internal Server Error

Error responses include a descriptive message:

```json
{
  "error": "Product sku_999 not found"
}
```

## Rate Limiting

Currently, no rate limiting is implemented. In production, consider implementing:

- Per-API-key rate limits
- Per-IP rate limits
- Burst protection
- Sliding window algorithms

## Webhook Security

Webhooks use HMAC-SHA256 signatures for verification:

```
X-Acp-Signature: t=1234567890,v1=<hex_digest>
```

Where `v1` is the HMAC-SHA256 digest of the request body using the `ACP_WEBHOOK_SECRET`.

## Data Models

### Money
```json
{
  "value": 119.99,
  "currency": "USD"
}
```

### Address
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "company": "Acme Corp",
  "address1": "123 Main St",
  "address2": "Suite 100",
  "city": "San Francisco",
  "state": "CA",
  "postal_code": "94102",
  "country": "US",
  "phone": "+1-555-123-4567"
}
```

### Payment
```json
{
  "method": "card",
  "token": "tok_test_123",
  "card": {
    "number": "4242424242424242",
    "exp_month": 12,
    "exp_year": 2025,
    "cvc": "123"
  }
}
```

## Examples

See the `/examples/` directory for complete test scripts demonstrating all API functionality.

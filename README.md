# ACP Gateway – Reference Implementation

A compact gateway that exposes Agentic Commerce Protocol (ACP) endpoints for LLM agents and merchant systems.

About ACP: https://github.com/agentic-commerce-protocol

## Features

- 🧠 **LLM-facing ACP endpoints** (`/acp/v1/*`)
- 🏪 **Merchant onboarding** + product feed ingest
- 🧪 **Sandbox mode** for safe testing
- 📊 **Audit logs** for all operations
- 🔐 **API key auth** + HMAC webhook verification
- 💾 **In-memory storage** (easily swappable for Redis/DB)

## Quick Start

### FastAPI (Python)

```bash
cd fastapi
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
export ACP_API_KEYS="test_key_123,another_key"
export ACP_WEBHOOK_SECRET="whsec_123"
uvicorn main:app --reload --port 8080
```

### Express (TypeScript)

```bash
cd express
npm install
export ACP_API_KEYS="test_key_123,another_key"
export ACP_WEBHOOK_SECRET="whsec_123"
npm run dev
```

### Test Both Implementations

```bash
./run-tests.sh
```

## API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/healthz` | Health check |
| `GET` | `/acp/v1/products` | List/search products |
| `POST` | `/acp/v1/intent` | Handle LLM intents (add to cart, etc.) |
| `POST` | `/acp/v1/checkout` | Create checkout session |
| `POST` | `/merchant/register` | Register new merchant |
| `POST` | `/merchant/feed` | Upload product feed |
| `POST` | `/webhooks/acp` | Receive webhooks |

## Authentication

- **API Requests**: `Authorization: Bearer <API_KEY>`
- **Sandbox Mode**: `X-ACP-Sandbox: true` header or `?sandbox=1`
- **Webhooks**: `X-Acp-Signature: t=<ts>,v1=<hex>` (HMAC-SHA256)

## Example Usage

### List Products
```bash
curl -H "Authorization: Bearer test_key_123" \
     http://localhost:8080/acp/v1/products
```

### Add to Cart
```bash
curl -X POST \
     -H "Authorization: Bearer test_key_123" \
     -H "Content-Type: application/json" \
     -d '{
       "type": "acp.intent",
       "actor": "llm",
       "payload": {
         "action": "add_to_cart",
         "items": [{"product_id": "sku_123", "quantity": 2}],
         "context": {"session_id": "session_123"}
       }
     }' \
     http://localhost:8080/acp/v1/intent
```

### Checkout
```bash
curl -X POST \
     -H "Authorization: Bearer test_key_123" \
     -H "Content-Type: application/json" \
     -d '{
       "cart_id": "cart_123",
       "email": "buyer@example.com",
       "payment": {"method": "card", "token": "tok_123"}
     }' \
     http://localhost:8080/acp/v1/checkout
```

## Project Structure

```
acp-gateway/
├── fastapi/           # FastAPI implementation
│   ├── main.py
│   ├── requirements.txt
│   └── env.example
├── express/           # Express/TypeScript implementation
│   ├── src/index.ts
│   ├── package.json
│   ├── tsconfig.json
│   └── env.example
├── schemas/           # JSON schemas for ACP protocol
│   ├── acp-intent.schema.json
│   ├── acp-intent-result.schema.json
│   ├── acp-checkout.schema.json
│   └── acp-checkout-response.schema.json
├── examples/          # Test scripts
│   ├── test-fastapi.py
│   ├── test-express.js
│   └── package.json
├── API.md            # Detailed API documentation
└── run-tests.sh      # Automated test runner
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `ACP_API_KEYS` | Comma-separated API keys | `test_key_123,another_key` |
| `ACP_WEBHOOK_SECRET` | Webhook HMAC secret | `whsec_123` |
| `PORT` | Server port (Express only) | `8080` |

## ACP Protocol Schemas

### Intent (LLM → Gateway)
```json
{
  "type": "acp.intent",
  "actor": "llm",
  "payload": {
    "action": "add_to_cart",
    "items": [
      { "product_id": "sku_123", "variant_id": "var_9", "quantity": 1 }
    ],
    "context": { "session_id": "abc", "locale": "en-US" }
  }
}
```

### Intent Result (Gateway → LLM)
```json
{
  "type": "acp.intent.result",
  "ok": true,
  "sandbox": true,
  "cart": {
    "id": "cart_001",
    "items": [...],
    "subtotal": {"value": 119.99, "currency": "USD"}
  },
  "next": {"action": "checkout", "endpoint": "/acp/v1/checkout"}
}
```

### Checkout Response
```json
{
  "ok": true,
  "sandbox": true,
  "order_id": "ord_123",
  "payment_intent_status": "requires_action",
  "redirect_url": "https://merchant.example/checkout/ord_123"
}
```

## Development

### Running Tests
```bash
# Test FastAPI
python examples/test-fastapi.py

# Test Express
node examples/test-express.js

# Test both (if servers are running)
./run-tests.sh
```

### Code Structure
- **FastAPI**: Uses Pydantic models, FastAPI decorators, and async/await
- **Express**: Uses TypeScript, Zod validation, and middleware patterns
- **Both**: Share identical API contracts and behavior

## Production Considerations

### Security
- Replace default API keys with strong, unique keys
- Use HTTPS in production
- Implement rate limiting and input validation
- Verify webhook signatures

### Scalability
- Replace in-memory storage with Redis/PostgreSQL
- Add connection pooling and caching
- Implement proper session management

### Monitoring
- Add structured logging and metrics
- Set up health checks and error tracking

## Documentation

- **[API.md](./API.md)** - Complete API reference with examples
- **[Schemas](./schemas/)** - JSON schemas for validation
- **[Examples](./examples/)** - Test scripts and usage examples

## License

MIT License - see [LICENSE](./LICENSE) file for details.

# ACP Gateway – Reference Implementation

A compact gateway that exposes Agentic Commerce Protocol (ACP) endpoints for LLM agents and merchant systems.

About ACP: 

https://github.com/agentic-commerce-protocol

https://stripe.com/blog/developing-an-open-standard-for-agentic-commerce


## Features

- 🧠 **LLM-facing ACP endpoints** (`/acp/v1/*`)
- 🏪 **Merchant onboarding** + product feed ingest
- 🧪 **Sandbox mode** for safe testing
- 📊 **Audit logs** for all operations
- 🔐 **API key auth** + HMAC webhook verification
- 💾 **In-memory storage** (easily swappable for Redis/DB)
- 🚀 **3 Language Implementations**: Python (FastAPI), TypeScript (Express), Go (High Performance)

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

### Go (High Performance)

```bash
cd go
# Fix PATH if Go is not found (common on macOS)
export PATH="/usr/local/go/bin:$PATH"

# Install dependencies and build
make deps

# Set environment variables and run
export ACP_API_KEYS="test_key_123,another_key"
export ACP_WEBHOOK_SECRET="whsec_123"
export PORT=8082
make run
```

**Note**: If you get "command not found: go", add this to your shell profile (`~/.zshrc` or `~/.bashrc`):
```bash
export PATH="/usr/local/go/bin:$PATH"
```

### Test All Implementations

```bash
./run-tests.sh
```

## API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/healthz` | Health check |
| `GET` | `/docs` | API documentation (HTML) |
| `GET` | `/openapi.json` | OpenAPI specification |
| `GET` | `/redoc` | ReDoc documentation (FastAPI only) |
| `GET` | `/acp/v1/products` | List/search products |
| `POST` | `/acp/v1/intent` | Handle LLM intents (add to cart, etc.) |
| `POST` | `/acp/v1/checkout` | Create checkout session |
| `POST` | `/merchant/register` | Register new merchant |
| `POST` | `/merchant/feed` | Upload product feed |
| `POST` | `/webhooks/acp` | Receive webhooks |

## API Documentation

Each implementation provides interactive API documentation:

### FastAPI (Python)
- **Swagger UI**: http://localhost:8080/docs
- **ReDoc**: http://localhost:8080/redoc
- **OpenAPI JSON**: http://localhost:8080/openapi.json

### Express (TypeScript)
- **HTML Docs**: http://localhost:8081/docs
- **OpenAPI JSON**: http://localhost:8081/openapi.json

### Go
- **HTML Docs**: http://localhost:8082/docs
- **OpenAPI JSON**: http://localhost:8082/openapi.json

## Authentication

- **API Requests**: `Authorization: Bearer <API_KEY>`
- **Sandbox Mode**: `X-ACP-Sandbox: true` header or `?sandbox=1`
- **Webhooks**: `X-Acp-Signature: t=<ts>,v1=<hex>` (HMAC-SHA256)

## Example Usage

### List Products
```bash
# FastAPI (port 8080)
curl -H "Authorization: Bearer test_key_123" \
     http://localhost:8080/acp/v1/products

# Express (port 8081) 
curl -H "Authorization: Bearer test_key_123" \
     http://localhost:8081/acp/v1/products

# Go (port 8082)
curl -H "Authorization: Bearer test_key_123" \
     http://localhost:8082/acp/v1/products
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
         "items": [{"id": "sku_123", "quantity": 2}],
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
       "cart_id": "session_123",
       "buyer": {
         "first_name": "John",
         "last_name": "Doe", 
         "email": "buyer@example.com"
       },
       "payment": {"method": "card", "token": "tok_123"}
     }' \
     http://localhost:8080/acp/v1/checkout
```

## Project Structure

```
acp-gateway/
├── fastapi/           # FastAPI implementation (Python)
│   ├── main.py
│   ├── requirements.txt
│   └── env.example
├── express/           # Express/TypeScript implementation
│   ├── src/index.ts
│   ├── package.json
│   ├── tsconfig.json
│   └── env.example
├── go/               # Go implementation (High Performance)
│   ├── cmd/server/
│   ├── internal/
│   ├── test/
│   ├── go.mod
│   ├── Makefile
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
├── review-acp-alignment.py  # Schema validation tool
├── API.md            # Detailed API documentation
└── run-tests.sh      # Automated test runner
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `ACP_API_KEYS` | Comma-separated API keys | `test_key_123,another_key` |
| `ACP_WEBHOOK_SECRET` | Webhook HMAC secret | `whsec_123` |
| `PORT` | Server port | `8080` (FastAPI), `8081` (Express), `8082` (Go) |

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
# Test FastAPI (port 8080)
python examples/test-fastapi.py

# Test Express (port 8081)
node examples/test-express.js

# Test Go (port 8082)
cd go
export PATH="/usr/local/go/bin:$PATH"  # Fix PATH if needed
make test-api

# Test all implementations (if servers are running)
./run-tests.sh
```

### Schema Validation
```bash
# Check schema alignment with official ACP specification
python review-acp-alignment.py
```

### Code Structure
- **FastAPI**: Uses Pydantic models, FastAPI decorators, and async/await
- **Express**: Uses TypeScript, Zod validation, and middleware patterns
- **Go**: Uses structs, Gorilla Mux router, and strong typing
- **All**: Share identical API contracts and behavior

### Performance Comparison

| Feature | FastAPI (Python) | Express (TypeScript) | Go |
|---------|------------------|---------------------|----| 
| **Performance** | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Memory Usage** | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Type Safety** | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Development Speed** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Deployment** | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Ecosystem** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |

**Recommendations:**
- **Development/Prototyping**: FastAPI (Python)
- **Full-stack JavaScript**: Express (TypeScript)  
- **Production/High Load**: Go

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

## Troubleshooting

### Go Setup Issues

**"command not found: go"**
```bash
# Check if Go is installed
ls -la /usr/local/go/bin/go

# Add to PATH permanently
echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**"no such file or directory: go"**
```bash
# Install Go (macOS with Homebrew)
brew install go

# Or download from https://golang.org/dl/
```

**Build errors in Go**
```bash
# Clean and rebuild
cd go
make clean
make deps
make build
```

### Port Conflicts

If you get "address already in use" errors:
- **FastAPI**: Change port with `--port 8081` (instead of 8080)
- **Express**: Set `PORT=8083` environment variable
- **Go**: Set `PORT=8084` environment variable

## Documentation

- **[API.md](./API.md)** - Complete API reference with examples
- **[Schemas](./schemas/)** - JSON schemas for validation
- **[Examples](./examples/)** - Test scripts and usage examples
- **[Go Implementation](./go/README.md)** - Go-specific documentation

## License

MIT License - see [LICENSE](./LICENSE) file for details.

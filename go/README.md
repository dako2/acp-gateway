# ACP Gateway - Go Implementation

A high-performance Go implementation of the Agentic Commerce Protocol (ACP) Gateway.

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or later
- Make (optional, for using Makefile commands)

### Installation

1. **Clone and navigate to Go directory:**
   ```bash
   cd go
   ```

2. **Install dependencies:**
   ```bash
   make deps
   # or manually: go mod download
   ```

3. **Set environment variables:**
   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

4. **Run the server:**
   ```bash
   make run
   # or manually: go run ./cmd/server
   ```

The server will start on port 8082 by default.

## 🧪 Testing

### Test API Endpoints
```bash
make test-api
# or manually: go run ./test/test_go.go
```

### Run Unit Tests
```bash
make test
# or manually: go test ./...
```

## 🛠️ Development

### Available Make Commands
```bash
make build      # Build the application
make run        # Run the application
make test       # Run tests
make test-api   # Test API endpoints
make clean      # Clean build artifacts
make deps       # Download dependencies
make fmt        # Format code
make vet        # Run go vet
make dev-setup  # Setup development environment
```

### Project Structure
```
go/
├── cmd/server/           # Main application entry point
├── internal/
│   ├── handlers/         # HTTP request handlers
│   ├── middleware/       # Authentication & CORS middleware
│   └── models/           # Data models and structs
├── test/                 # Test scripts
├── go.mod               # Go module definition
├── Makefile             # Build automation
└── README.md            # This file
```

## 📡 API Endpoints

### Health Check
- `GET /healthz` - Server health status

### ACP Endpoints (Require API Key)
- `GET /acp/v1/products` - List products
- `POST /acp/v1/intent` - Process intent requests
- `POST /acp/v1/checkout` - Process checkout requests

### Merchant Endpoints (Require API Key)
- `POST /merchant/register` - Register merchant
- `POST /merchant/feed` - Upload product feed

### Webhook Endpoints (Require Signature)
- `POST /webhooks/acp` - Handle ACP webhooks

## 🔐 Authentication

### API Key Authentication
Include the API key in the Authorization header:
```
Authorization: Bearer your_api_key
```

### Webhook Signature Verification
Webhook requests must include a valid HMAC signature:
```
X-Acp-Signature: t=timestamp,v1=signature
```

## 🌍 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ACP_API_KEYS` | Comma-separated API keys | `test_key_123,another_key` |
| `ACP_WEBHOOK_SECRET` | Webhook signature secret | `whsec_123` |
| `PORT` | Server port | `8082` |
| `GO_ENV` | Environment mode | `development` |

## 🏗️ Production Build

### Docker Build
```bash
make prod-build
```

### Manual Production Build
```bash
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/acp-gateway-go ./cmd/server
```

## 🔧 Configuration

The server uses in-memory storage for demonstration purposes. In production, you would:

1. Replace in-memory storage with a database
2. Implement proper logging
3. Add metrics and monitoring
4. Configure proper TLS/SSL
5. Set up proper secret management

## 📊 Features

- ✅ **High Performance**: Built with Go for optimal performance
- ✅ **Type Safety**: Strong typing with Go structs
- ✅ **Middleware**: Authentication, CORS, and logging
- ✅ **Testing**: Comprehensive test suite
- ✅ **Documentation**: Auto-generated API docs
- ✅ **Production Ready**: Docker and production build support

## 🆚 Comparison with Other Implementations

| Feature | Go | FastAPI | Express |
|---------|----|---------|---------| 
| Performance | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| Type Safety | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ |
| Memory Usage | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| Deployment | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Format code: `make fmt`
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

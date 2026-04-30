# Stock Recommendation Backend - Documentation

## 🎯 Project Overview
A production-ready Go backend using Gorilla Mux for a fintech stock recommendation system. This backend is designed to integrate with a Python FastAPI microservice for ML predictions.

## 📁 Project Structure

```
stock-backend/
├── main.go                 # Entry point - server initialization
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── config/
│   └── config.go          # Configuration loader (env variables)
├── routes/
│   └── routes.go          # Router setup with all endpoints
├── handlers/
│   └── handlers.go        # HTTP request handlers (business logic)
├── models/
│   └── models.go          # Data structures and DTOs
└── middleware/
    └── middleware.go      # HTTP middleware (logging, CORS, recovery)
```

## 🚀 Quick Start

### Prerequisites
- Go 1.23 or higher
- Gorilla Mux dependency (installed via `go mod tidy`)

### Running the Server

```bash
cd stock-backend
go mod tidy          # Install dependencies
go run main.go       # Start the server
```

The server will start on `localhost:8080`

## 📊 API Endpoints

### 1. Health Check
**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "success",
  "message": "Server is running"
}
```

**Purpose:** Verify the server is running and responsive.

### 2. Stock Recommendation (Placeholder)
**Endpoint:** `POST /api/v1/recommend`

**Request Body:**
```json
{
  "market_history": [[...], [...]],  // 60x5 array
  "user_profile": [0.5, 0.8],        // risk_tolerance, holding_period
  "portfolio_weights": [0.1, 0.15, ...], // 10 elements
  "news_sentiment": [0.7, 0.3, 0.2]  // sentiment scores
}
```

**Response (Current Placeholder):**
```json
{
  "recommendation_score": 0.5,
  "confidence": 0.0,
  "status": "pending_ml_service"
}
```

**Future:** This will call the FastAPI ML microservice

## 🏗️ Architecture

### Clean Architecture Principles Applied
- **Separation of Concerns:** Each package has a single responsibility
- **Modularity:** Easy to add new features without modifying existing code
- **Testability:** Clear interfaces and dependencies
- **Extensibility:** Prepared for ML service integration

### Package Responsibilities

**main.go**
- Server initialization
- Configuration loading
- HTTP server startup

**config/**
- Loads server configuration from environment variables
- Provides default values for Host, Port, and Environment

**routes/**
- Defines all API endpoints
- Applies global middleware
- Sets up subrouters for API versioning

**handlers/**
- Contains HTTP request handlers
- Implements business logic
- Placeholder for ML service integration

**models/**
- Data transfer objects (DTOs)
- Request/Response structures
- Prepared for ML service data structures

**middleware/**
- Logging middleware (logs all requests)
- CORS middleware (allows cross-origin requests)
- Recovery middleware (catches panics gracefully)

## 🔗 Future ML Integration

The backend is designed to easily integrate with a FastAPI ML service:

### When ML Service is Ready:
1. Update `handlers.StockRecommendationHandler()` to call the FastAPI service
2. Add a `services/` package with ML client logic
3. Call the ML service at `http://localhost:8000/recommend` (or your server)
4. Transform and return the prediction

### Example Integration Code (Future):
```go
// In services/ml_client.go
func CallMLService(req models.StockRecommendationRequest) (*models.StockRecommendationResponse, error) {
    resp, err := http.Post(
        "http://localhost:8000/recommend",
        "application/json",
        // marshal request to JSON
    )
    // Parse response and return
}
```

## 🔧 Configuration

Environment variables:
- `SERVER_HOST` - Server host (default: `localhost`)
- `SERVER_PORT` - Server port (default: `8080`)
- `ENVIRONMENT` - Environment (default: `development`)

Example:
```bash
export SERVER_PORT=9000
export ENVIRONMENT=production
go run main.go
```

## 📝 Middleware Features

### Logging Middleware
Logs all incoming requests with method, URI, and remote address.

### CORS Middleware
Allows cross-origin requests from any origin (can be restricted in production).

### Recovery Middleware
Catches panics and returns a JSON error response instead of crashing.

## 🧪 Testing Health Check

```bash
# Test health endpoint
curl http://localhost:8080/health

# With pretty JSON output
curl http://localhost:8080/health | jq .

# Test 404
curl http://localhost:8080/invalid
```

## 📦 Dependencies

- **github.com/gorilla/mux** - HTTP router and dispatcher

Minimal dependencies for production-ready code.

## 🎯 Next Steps

1. ✅ Backend is running and healthy
2. **Integration Phase:** When the FastAPI ML model is ready
   - Add the trained model endpoint
   - Implement ML service client in `services/` package
   - Update `StockRecommendationHandler` to call ML service
3. **Frontend Integration:** Connect from FRONTEND-MAIN to this backend
   - Update frontend API client to call `POST /api/v1/recommend`
   - Handle ML predictions in UI

## 📋 Code Quality

- ✅ Clean architecture
- ✅ Modular structure
- ✅ Proper error handling
- ✅ Logging for all requests
- ✅ CORS support
- ✅ Panic recovery
- ✅ Production-ready
- ✅ Extensible for future features

## 🚨 Important Notes

- The server listens on `localhost:8080` (change `SERVER_PORT` env var to modify)
- CORS is open to all origins (restrict in production)
- ML service integration is prepared but not yet implemented
- All responses are JSON formatted

---

**Status:** ✅ Ready for Frontend Integration and ML Service Connection

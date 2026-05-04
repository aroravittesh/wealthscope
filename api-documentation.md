# AI Finance Service — API Documentation

## Overview

This API provides a finance-scoped AI service with chatbot interaction, sentiment analysis, stock data retrieval, portfolio risk assessment, and feedback collection. The backend is informational and does not provide personalized financial advice.

---

## Endpoints

### 1. Health Check

**`GET /health`**

Checks whether the AI service is running.

**Example Response:**
```json
{
  "success": true,
  "message": "AI Service running",
  "data": { "status": "ok" },
  "error": null
}
```

---

### 2. Chat

**`POST /chat`**

Primary chatbot interaction endpoint. Supports improved intent handling, entity extraction, retrieval-augmented grounding, optional live web context for time-sensitive queries, follow-up context carryover, and natural paragraph-based answers.

**Request Body:**
```json
{
  "message": "Compare Apple and Microsoft",
  "session_id": "demo-session-1"
}
```

**Example Response:**
```json
{
  "success": true,
  "message": "Chat response generated successfully",
  "data": {
    "response": "Apple and Microsoft are both large-cap technology companies...",
    "follow_ups": [
      "Should I add risk and volatility context?",
      "Want a quick fundamentals comparison too?",
      "Want latest news for both names as well?"
    ]
  },
  "error": null
}
```

---

### 3. Clear Chat Session

**`DELETE /chat/session/:session_id`**

Clears stored session state for the given session ID. Resets chatbot memory, follow-up context, and conversation history for the selected session.

---

### 4. Intent Detection

**`POST /intent`**

Detects the user's finance-related intent. Supported categories include: stock price, market news, risk analysis, portfolio help, and general finance concepts.

**Request Body:**
```json
{
  "message": "What is the latest news on Tesla?"
}
```

---

### 5. Sentiment Analysis

**`POST /sentiment`**

Analyzes sentiment for finance-related text input.

**Request Body:**
```json
{
  "text": "Tesla reported stronger-than-expected growth."
}
```

---

### 6. News Sentiment

**`GET /news-sentiment/:symbol`**

Returns aggregated sentiment for recent stock-related news articles for a given ticker.

**Response Fields:**

| Field | Description |
|---|---|
| `symbol` | Stock ticker symbol |
| `overall_sentiment` | Aggregated sentiment rating |
| `confidence` | Confidence score |
| `article_count` | Number of articles analyzed |
| `top_positive_article` | Highest positive article reference |
| `top_negative_article` | Highest negative article reference |
| `summary` | Plain-language sentiment summary |

---

### 7. Stock Quote

**`GET /quote/:symbol`**

Fetches live or provider-backed quote information for a stock ticker.

---

### 8. Company Overview

**`GET /company/:symbol`**

Fetches company-level overview or fundamentals for a given ticker.

---

### 9. Market News

**`GET /news/:symbol`**

Fetches recent market or stock-related news for a given ticker.

---

### 10. Portfolio Risk

**`POST /risk`**

Calculates portfolio risk using backend scoring logic.

**Request Body:**
```json
{
  "holdings": [
    { "symbol": "TSLA", "allocation": 0.5, "beta": "2.0" },
    { "symbol": "NVDA", "allocation": 0.5, "beta": "1.8" }
  ]
}
```

---

### 11. Risk Drift Prediction

**`POST /predict/risk-drift`**

Estimates whether a portfolio is drifting away from the intended risk profile.

**Request Body:**
```json
{
  "holdings": [
    { "symbol": "TSLA", "allocation": 0.5, "beta": "2.0" },
    { "symbol": "NVDA", "allocation": 0.5, "beta": "1.8" }
  ],
  "target_risk": "MEDIUM"
}
```

**Response Fields:** `drift_level`, `score`, `explanation`, and optional explainability fields or top drivers.

---

### 12. Portfolio Explanation

**`POST /portfolio/explain`**

Returns a human-readable explanation of portfolio risk, concentration, diversification, and alignment. Remains neutral and non-advisory; supports both dashboard and chatbot use cases.

**Response Fields:**

| Field | Description |
|---|---|
| `summary` | Overall portfolio summary |
| `top_risks` | Key identified risk factors |
| `concentration_warning` | Concentration alerts if applicable |
| `diversification_note` | Diversification observations |
| `risk_alignment` | Alignment to target risk profile |
| `neutral_guidance` | Non-advisory contextual guidance |

---

### 13. Portfolio Summarize

**`POST /portfolio/summarize`**

Returns a concise structured summary of holdings and portfolio composition.

---

### 14. Portfolio Changes

**`POST /portfolio/changes`**

Describes changes in portfolio profile relative to a prior portfolio snapshot.

---

### 15. Multi-Stock Comparison

**`POST /compare`**

Compares multiple stocks in a structured format. Useful for frontend comparison cards and chatbot grounding.

**Request Body:**
```json
{
  "symbols": ["AAPL", "MSFT"]
}
```

**Response Fields:** `comparisons`, `summary`

---

### 16. Feedback Collection

**`POST /feedback`**

Captures user feedback for chatbot or AI-generated outputs. Collects quality signals for future improvement and logs whether a response was helpful, unclear, or incorrect.

**Example Request:**
```json
{
  "session_id": "demo-session-1",
  "query": "Compare Apple and Microsoft",
  "response_type": "chat",
  "feedback": "not_helpful",
  "reason": "The answer was too generic"
}
```

---

## Sprint 4 Backend Enhancements

The following improvements are reflected in the current API layer:

- Suggested prompt and follow-up aware chatbot flow
- Improved response formatting and readability
- Follow-up context carryover across turns
- Finance-specific sentiment support
- Better entity resolution for company and ticker references
- Improved retrieval reranking and grounding
- Live web-context integration for time-sensitive queries
- Explainability support for AI decisions
- Structured logging for chatbot routing and fallback behavior
- Reusable validation helpers
- Centralized configuration and environment handling
- Feedback endpoint for future model improvement

---

## Notes

- The backend is **finance-scoped and informational** — it does not provide personalized financial advice.
- Live web context is used in a controlled, backend-driven way for time-sensitive queries.
- Retrieval can combine internal finance knowledge, CSV-backed Q&A, market and news provider data, and optional live web context.
- Session handling supports follow-up continuity for a better conversational experience.

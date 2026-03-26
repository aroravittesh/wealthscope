# 📅 Sprint 2 – Holdings, Integration & Testing

## 🎯 Work Completed

### Backend

* Implemented **Holdings module** (Add, Get, Delete)
* Integrated holdings with portfolio system
* Followed clean architecture (Repository → Service → Handler)
* Secured endpoints using JWT authentication
* Added API routes for holdings

---

### Frontend

#### Portfolio Management UI

* Designed and implemented **Portfolio interface**
* Enabled users to:

  * Create and view portfolios
  * Navigate into individual portfolios
  * Manage holdings within each portfolio

#### Dynamic Routing

* Implemented route:

  ```
  /portfolio/:portfolioId/holdings
  ```
* Dynamically fetched data based on selected portfolio

#### Holdings Management

* Built UI to:

  * Add holdings
  * Delete holdings
  * View asset details
* Ensured instant UI updates after actions

#### Data Integration

* Connected frontend with backend APIs
* Displayed real-time portfolio and holdings data
* Maintained consistent UI state

---

### Frontend Fixes & Improvements

* Fixed duplicate navbar issue
* Resolved Angular compile errors (missing imports in `portfolio.service.ts`)
* Improved routing and component structure
* Ensured proper state handling

---

### Request Handling Fix

* Normalized request payload:

  * `symbol` → uppercase
  * `asset_type` → (`STOCK`, `CRYPTO`, `ETF`)
* Updated validation to prevent invalid inputs

---

### Error Handling & UX

* Display backend error messages in UI
* Improved debugging for failed API requests

---

### AI Recommendations (Frontend)

* Developed **AI Recommendations dashboard**

* Included:

  * Preferences panel (user input)
  * Recommendations panel (output display)

* User inputs:

  * Portfolio tickers (AAPL, TSLA, etc.)
  * Risk profile
  * Investment horizon

* Displayed results:

  * Asset names
  * Confidence scores
  * BUY / HOLD / SELL decisions
  * Expected returns
  * Model explanation

---

### ML Integration

* Sent user inputs to backend APIs
* Triggered ML processing via backend
* Received and rendered predictions dynamically
* Displayed scores, decisions, and returns clearly

---

## 🧪 Testing

### Frontend (Angular Unit Tests)

* `holdings.component.spec.ts`

  * Validates input normalization
  * Ensures correct payload sent

* `portfolio.service.spec.ts`

  * Maps backend response to frontend model
  * Handles field transformation and parsing

* Fixed:

  * `login.spec.ts`
  * `signup.spec.ts`
  * `app.component.spec.ts`

---

### Frontend (Cypress E2E)

* `login-form.cy.js`

  * Tests login UI interaction
  * Validates form behavior

---

### Backend Testing

* Service layer unit tests:

  * AddHolding
  * GetHoldings
  * DeleteHolding

* Handler testing using `httptest`

* Manual API testing using curl

---

### ML Testing

* Unit tests for:

  * intent detection
  * sentiment analysis
  * risk scoring
* Verified correct outputs for different inputs

---

## 📡 Backend API Documentation

### Base URL

```id="x9n2wk"
http://localhost:8080/api
```

### Auth

```id="q4m7ld"
POST /auth/register  
POST /auth/login
```

### Portfolio

```id="p2v8rs"
POST /portfolios  
GET /portfolios
```

### Holdings

```id="c7k3zn"
POST /holdings  
GET /holdings/{portfolio_id}  
DELETE /holdings/{id}
```

### Additional Endpoints

```id="h5w1yx"
GET /quote/:symbol  
GET /company/:symbol  
GET /news/:symbol  
POST /intent  
POST /sentiment  
POST /risk  
DELETE /chat/session/:session_id
```

---

## 🚀 Outcome

* Fully functional holdings feature integrated end-to-end
* Strong frontend-backend integration
* AI dashboard integrated with backend
* Added unit and E2E testing coverage

---

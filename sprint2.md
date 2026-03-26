# 📅 Sprint 2 – Holdings, Integration & Testing

## 🎯 Work Completed

### Backend

* Implemented **Holdings module** (Add, Get, Delete)
* Integrated holdings with portfolio system
* Followed clean architecture:

  * Repository layer (DB operations)
  * Service layer (business logic)
  * Handler layer (API endpoints)
* Secured endpoints using JWT authentication

### Frontend

* Integrated holdings feature end-to-end
* Added route: `/portfolio/:portfolioId/holdings`
* Displayed holdings data in table (fetched from backend)
* Enabled add/delete operations with UI updates
* Connected frontend with backend for:

  * Login
  * Portfolio
  * Holdings
* Fixed UI issues:

  * Duplicate navbar removed
  * Routing and state fixes

### API Compatibility Fix

* Normalized request payload:

  * `symbol` → uppercase
  * `asset_type` → enum format (`STOCK`, `CRYPTO`, `ETF`)
* Updated form validation to prevent invalid inputs

### Error Handling

* Display backend error messages in UI
* Improved debugging for failed API requests

---

## 🧪 Frontend Testing

### Angular Unit Tests

* `holdings.component.spec.ts`

  * Tests input normalization (uppercase symbol & asset_type)
  * Verifies correct payload sent to backend

* `portfolio.service.spec.ts`

  * Tests mapping of backend response:

    * `portfolio_id → portfolioId`
    * `asset_type → assetType`
  * Validates number casting and date parsing

* Fixed existing tests:

  * `login.spec.ts`
  * `signup.spec.ts`
  * `app.component.spec.ts`

### Cypress E2E Test

* `login-form.cy.js`

  * Visits `/auth/login`
  * Inputs email & password
  * Verifies submit button enables when form is valid

---

## 🧪 Backend Testing

* Service layer unit tests using mock repository:

  * AddHolding
  * GetHoldings
  * DeleteHolding

* Handler tests using `httptest`

* API testing using curl:

  * Register, Login
  * Portfolio creation
  * Holdings CRUD operations

---

## 📡 Backend API Documentation

### Base URL

```
http://localhost:8080/api
```

### Auth

#### Register

```
POST /auth/register
```

Body:

```
{
  "email": "string",
  "password": "string",
  "risk_preference": "LOW | MEDIUM | HIGH"
}
```

#### Login

```
POST /auth/login
```

---

### Portfolio

#### Create Portfolio

```
POST /portfolios
Authorization: Bearer <token>
```

#### Get Portfolios

```
GET /portfolios
Authorization: Bearer <token>
```

---

### Holdings

#### Add / Update Holding

```
POST /holdings
Authorization: Bearer <token>
```

Body:

```
{
  "portfolio_id": "uuid",
  "symbol": "AAPL",
  "asset_type": "STOCK",
  "quantity": 10,
  "avg_price": 150
}
```

#### Get Holdings

```
GET /holdings/{portfolio_id}
Authorization: Bearer <token>
```

#### Delete Holding

```
DELETE /holdings/{id}
Authorization: Bearer <token>
```

---

## 🚀 Outcome

* Fully functional **Holdings feature (end-to-end)**
* Stable frontend-backend integration
* Improved UI and error handling
* Added unit and E2E test coverage

---



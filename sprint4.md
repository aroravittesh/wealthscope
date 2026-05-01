# WealthScope Feature Documentation

## Team Members and Roles

- **Vittesh Arora** – Backend (RBAC/admin APIs, assets, portfolio snapshots, audit logs, snapshot compare/trend)
- **Raghav Gupta** – Frontend Admin Dashboard (admin operations + tests)
- **Rishithaa Maligireddy** – Frontend Development and AI Recommender UI Integration
- **Ansh Jain** – AI Service Backend and Intelligence Features

---

## Features Covered

1. **Admin Audit Logs**
2. **Snapshot Compare & Trend Insights**

This document is written with separate, detailed sections for **Backend** and **Frontend** for both features.

---

# 1) Admin Audit Logs

## 1.1 Backend

### Objective
Track sensitive/admin actions in a persistent audit trail so the system can answer:
- who performed an action,
- what changed,
- when it changed,
- which entity was affected.

### Backend implementation status
Implemented.

### Data model

Audit logs are stored in `audit_logs` with the following fields:
- `id` (UUID primary key)
- `actor_user_id` (UUID, nullable, FK to users)
- `action` (text, required)
- `entity_type` (text, required)
- `entity_id` (UUID, nullable)
- `before_json` (text, optional)
- `after_json` (text, optional)
- `created_at` (timestamp)

### SQL (Supabase)
Use this once in Supabase SQL editor:

```sql
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id UUID,
    before_json TEXT,
    after_json TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_user_id ON audit_logs(actor_user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_type ON audit_logs(entity_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
```

If needed:
```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
```

### Logged actions currently covered
- `USER_ROLE_UPDATED`
- `ASSET_CREATED`
- `ASSET_UPDATED`
- `ASSET_DELETED`

### API endpoint
- `GET /api/admin/audit-logs?limit=50`
  - auth required
  - admin role required
  - `limit` optional, default `50`, max `200`

### Response shape (example)
```json
[
  {
    "id": "uuid",
    "actor_user_id": "uuid",
    "action": "ASSET_UPDATED",
    "entity_type": "asset",
    "entity_id": "uuid",
    "before_json": "{\"symbol\":\"TSLA\"}",
    "after_json": "{\"symbol\":\"TSLA\",\"name\":\"Tesla, Inc.\"}",
    "created_at": "2026-05-01T00:00:00Z"
  }
]
```

### Backend testing checklist
- Admin can fetch logs (`200`)
- Non-admin gets `403`
- Invalid/malformed token gets `401`
- `limit` validation returns `400` for invalid values
- Role update and asset mutations append new log records

---

## 1.2 Frontend

### Objective
Provide an admin-facing “Activity Log” screen for visibility and traceability.

### Suggested UI scope
- New admin tab/page: **Audit Logs**
- Table columns:
  - Time
  - Actor
  - Action
  - Entity Type
  - Entity ID
  - Before
  - After
- Filter controls:
  - action filter
  - entity type filter
  - actor/email filter
  - date range
- Pagination / limit selector
- Quick JSON expand/collapse for before/after

### Suggested frontend service methods
- `getAuditLogs(limit: number)`
- later optional: server-side filters once backend supports them

### UX notes
- Default newest first
- Show clean badge styles by action type
- Provide copy-to-clipboard for entity IDs
- Friendly empty state: “No admin activities yet.”

---

# 2) Snapshot Compare & Trend Insights

## 2.1 Backend

### Objective
Make snapshots analytically useful by enabling:
- pairwise comparison (Then vs Now)
- chart-ready trend data over time

### Backend implementation status
Implemented.

### Existing snapshot base
Already used:
- `POST /api/portfolios/{id}/snapshots`
- `GET /api/portfolios/{id}/snapshots`

### Compare endpoint
- `GET /api/portfolios/{id}/snapshots/compare?from=<snapshotId>&to=<snapshotId>`

Validations:
- auth required
- ownership check for portfolio
- `from` and `to` must both be present and different
- both snapshot IDs must belong to same portfolio

### Compare response includes
- snapshot metadata:
  - `portfolio_id`, `from_id`, `to_id`, `from_at`, `to_at`
- metric deltas:
  - `total_value_delta`
  - `total_invested_delta`
  - `profit_loss_delta`
  - `diversification_delta`
  - `volatility_delta`
- allocation drift:
  - symbol-wise changes in percent and value
  - supports added/removed symbols between snapshots

### Trend endpoint
- `GET /api/portfolios/{id}/snapshots/trend?limit=20`

Returns chronological points:
- `snapshot_id`
- `created_at`
- `total_portfolio_value`
- `total_invested`
- `total_profit_loss`
- `diversification`
- `volatility`

### Backend testing checklist
- Compare: valid IDs returns `200`
- Compare: missing params returns `400`
- Compare: same from/to returns `400`
- Compare: invalid/missing snapshot returns `404`
- Trend: default limit works
- Trend: invalid limit returns `400`
- Trend: ownership enforced (`403` for non-owner)

---

## 2.2 Frontend

### Objective
Expose real “Then vs Now” insights on analytics page for users.

### Suggested UI scope

#### A) Snapshot Compare panel
- Two dropdowns:
  - **From Snapshot**
  - **To Snapshot**
- Compare button
- Cards for:
  - Total Value change
  - P/L change
  - Diversification change
  - Volatility change
- Allocation drift table:
  - Symbol
  - From %
  - To %
  - Delta %
  - From value
  - To value
  - Delta value

#### B) Trend mini chart
- Small line chart for:
  - total portfolio value
  - optionally diversification and volatility toggles
- X-axis: snapshot timestamps
- Y-axis: selected metric

### Suggested frontend service methods
- `compareSnapshots(portfolioId, fromSnapshotId, toSnapshotId)`
- `getSnapshotTrend(portfolioId, limit?)`

### UX notes
- Disable compare if fewer than 2 snapshots
- Default selection:
  - from = previous snapshot
  - to = latest snapshot
- Highlight positive/negative deltas with color coding
- Keep “Live view” and “Snapshot view” context labels visible

---

# 3) End-to-End Backend Curl Commands

These commands cover login + audit logs + asset create/update (and audit verify) + snapshot compare.

## 3.1 Login

```bash
BASE="http://localhost:8080/api"
EMAIL="vittesharora@test.com"
PASS="test123"

LOGIN_RES=$(curl -sS -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}")

ACCESS_TOKEN=$(echo "$LOGIN_RES" | jq -r '.access_token')
echo "$LOGIN_RES" | jq
```

## 3.2 Fetch admin audit logs

```bash
curl -sS "$BASE/admin/audit-logs?limit=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq
```

## 3.3 Create new stock in admin assets

```bash
SYM="ADM$(date +%s | tail -c 5)"

CREATE_RES=$(curl -sS -X POST "$BASE/admin/assets" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"symbol\":\"$SYM\",\"name\":\"Admin Test Asset\",\"asset_type\":\"STOCK\"}")

ASSET_ID=$(echo "$CREATE_RES" | jq -r '.id')
echo "$CREATE_RES" | jq
echo "$ASSET_ID"
```

## 3.4 Update stock and verify logs again

```bash
curl -sS -X PUT "$BASE/admin/assets/$ASSET_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"symbol\":\"$SYM\",\"name\":\"Admin Test Asset Updated\",\"asset_type\":\"STOCK\"}" | jq

curl -sS "$BASE/admin/audit-logs?limit=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq
```

## 3.5 Snapshot compare (already existing snapshots)

```bash
PORTFOLIO_ID=$(curl -sS "$BASE/portfolios" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq -r '.[0].id')

SNAPSHOT_1=$(curl -sS "$BASE/portfolios/$PORTFOLIO_ID/snapshots" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq -r '.[0].id')

SNAPSHOT_2=$(curl -sS "$BASE/portfolios/$PORTFOLIO_ID/snapshots" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq -r '.[1].id')

curl -sS "$BASE/portfolios/$PORTFOLIO_ID/snapshots/compare?from=$SNAPSHOT_2&to=$SNAPSHOT_1" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq
```

Optional trend check:
```bash
curl -sS "$BASE/portfolios/$PORTFOLIO_ID/snapshots/trend?limit=20" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq
```

---

# 4) Acceptance Criteria (Combined)

## Admin Audit Logs
- Admin actions generate log records with actor/action/entity and change payloads.
- Admin can fetch logs via API.
- Non-admin users cannot access admin log endpoint.

## Snapshot Compare & Trend
- User can compare any two snapshots in same portfolio.
- Response includes core deltas and allocation drift.
- User can fetch trend points for chart rendering.
- Ownership checks are enforced.

---

# 5) Notes

- Backend parts for both features are implemented.
- Frontend sections in this document represent detailed implementation scope/UI contract and can be used as execution guide.
- Ensure Supabase schema includes both:
  - `portfolio_snapshots` table
  - `audit_logs` table

---

# Appendix A — AI Service API Documentation (Integrated)

## Overview

This API provides a finance-scoped AI service with chatbot interaction, sentiment analysis, stock data retrieval, portfolio risk assessment, and feedback collection. The backend is informational and does not provide personalized financial advice.

## Endpoints

### 1. Health Check
**`GET /health`**

### 2. Chat
**`POST /chat`**

Primary chatbot endpoint with intent handling, entity extraction, retrieval-augmented grounding, optional live web context, and follow-up context carryover.

### 3. Clear Chat Session
**`DELETE /chat/session/:session_id`**

### 4. Intent Detection
**`POST /intent`**

### 5. Sentiment Analysis
**`POST /sentiment`**

### 6. News Sentiment
**`GET /news-sentiment/:symbol`**

Returns aggregated sentiment with confidence and article-level highlights.

### 7. Stock Quote
**`GET /quote/:symbol`**

### 8. Company Overview
**`GET /company/:symbol`**

### 9. Market News
**`GET /news/:symbol`**

### 10. Portfolio Risk
**`POST /risk`**

### 11. Risk Drift Prediction
**`POST /predict/risk-drift`**

### 12. Portfolio Explanation
**`POST /portfolio/explain`**

### 13. Portfolio Summarize
**`POST /portfolio/summarize`**

### 14. Portfolio Changes
**`POST /portfolio/changes`**

### 15. Multi-Stock Comparison
**`POST /compare`**

### 16. Feedback Collection
**`POST /feedback`**

## Sprint 4 Backend Enhancements Reflected

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

## Notes

- The backend is **finance-scoped and informational** and does not provide personalized financial advice.
- Live web context is used in a controlled backend-driven way for time-sensitive queries.
- Retrieval can combine internal finance knowledge, CSV-backed Q&A, market/news provider data, and optional live web context.
- Session handling supports follow-up continuity for better conversational experience.

---

# Appendix B — Sprint Documentation (Frontend)

## Sprint Goal

Add meaningful snapshot analytics in the frontend so users can compare historical portfolio states, see trend progression, and understand allocation drift over time.

## Features Implemented

### 1) Snapshot Compare (Then vs Now)

- Added compare controls on analytics page to select two snapshots:
  - Then (from snapshot)
  - Now (to snapshot)
- Integrated compare API response into frontend and rendered delta cards for:
  - Total Value Change
  - Profit/Loss Change
  - Diversification Change
  - Volatility Change
- Added validation to prevent comparing the same snapshot.

**Key files**
- `Frontend/src/app/features/reporting-analytics.component.ts`
- `Frontend/src/app/services/portfolio.service.ts`
- `Frontend/src/app/models/index.ts`

### 2) Snapshot Trend Insights

- Integrated snapshot trend endpoint for chart-ready points.
- Added mini trend chart (SVG polyline) on analytics page for total portfolio value across snapshots.
- Added start/end date labels and point count for readability.

**Key files**
- `Frontend/src/app/features/reporting-analytics.component.ts`
- `Frontend/src/app/services/portfolio.service.ts`
- `Frontend/src/app/models/index.ts`

### 3) Allocation Drift Insights

- Added “Allocation Drift Insights” table to show top movers between snapshots.
- Displays:
  - Symbol
  - Then %
  - Now %
  - Drift %
  - Value Change
- Sorted by highest absolute drift to surface biggest changes first.

**Key file**
- `Frontend/src/app/features/reporting-analytics.component.ts`

## Unit Test Documentation

### A) Portfolio Service Tests
**File:** `Frontend/src/app/services/portfolio.service.spec.ts`

Added/covered tests:
- `getPortfolioSnapshotCompare should map delta payload`
  - Verifies compare endpoint mapping:
    - IDs, timestamps
    - delta metrics
    - allocation drift rows
- `getPortfolioSnapshotTrend should map trend points`
  - Verifies trend endpoint mapping:
    - portfolio id
    - snapshot points
    - date/value metrics

### B) Reporting Analytics Component Tests
**File:** `Frontend/src/app/features/reporting-analytics.component.spec.ts`

Added tests:
- `should initialize analytics with first portfolio and load trend`
- `runSnapshotCompare should set error when snapshot ids are invalid`
- `runSnapshotCompare should populate compare response on success`
- `topAllocationDriftRows should sort by largest absolute drift`
- `trendPolylinePoints should return chart points for trend data`

### C) Existing Admin Test Stabilization
**File:** `Frontend/src/app/features/admin-dashboard.component.spec.ts`

- Made one date-dependent assertion deterministic by fixing sample date in test data, to keep suite reliable.

## Cypress Test Documentation

### Analytics E2E Test
**File:** `Frontend/cypress/e2e/analytics-snapshot-compare.cy.js`

Test flow:
- Visits `/analytics` with mocked authenticated user session.
- Intercepts/mocks:
  - portfolios list
  - portfolio summary
  - snapshots list
  - trend endpoint
  - compare endpoint
- Clicks Compare Snapshots.
- Verifies UI outputs:
  - compare cards visible
  - trend section visible
  - allocation drift insights visible

## Test Execution Commands

### Unit Tests
```bash
cd /Users/raghhavv03/wealthscope/Frontend
npx ng test --no-watch --browsers=ChromeHeadless
```

### Cypress Tests
```bash
cd /Users/raghhavv03/wealthscope/Frontend
npm start
```

In another terminal:
```bash
cd /Users/raghhavv03/wealthscope/Frontend
npm run cypress:run
```

## Outcome

This sprint upgraded analytics from static snapshot viewing to actionable insight:
- quick Then-vs-Now deltas,
- temporal trend visibility,
- allocation drift explainability,
- with automated frontend coverage via unit and Cypress tests.

# Sprint 3 Documentation – WealthScope

## Team Members and Roles

- **Vittesh Arora** – Backend (RBAC/admin APIs, assets, portfolio snapshots)
- **Raghav Gupta** – Frontend Admin Dashboard (admin operations + tests)
- **Rishithaa Maligireddy** – Frontend Development and AI Recommender UI Integration
- **Ansh Jain** – AI Service Backend and Intelligence Features


---

## Sprint 3 – Vittesh Arora’s Work (Backend)

### Overview
Implemented missing backend pieces for **RBAC/admin controls**, **asset master management**, and **portfolio reporting via stored snapshots** (email reporting was removed as requested).

### 1) RBAC / Admin controls

- **JWT role support**
  - Access tokens include a `role` claim from `users.role`
  - `AuthMiddleware` injects `user_id` and `role` into request context
- **Role enforcement middleware**
  - `RequireRole("ADMIN")` blocks non-admin users with `403 forbidden`

### 2) Admin APIs (ADMIN-only)

- **User management**
  - `GET /api/admin/users` → list all users (safe/public fields only)
  - `PATCH /api/admin/users/{id}/role` → update role (`USER` or `ADMIN`)
- **Asset master management**
  - `GET /api/admin/assets` → list assets
  - `POST /api/admin/assets` → create asset
  - `PUT /api/admin/assets/{id}` → update asset
  - `DELETE /api/admin/assets/{id}` → delete asset

### 3) Portfolio snapshot reporting (owner-only)

- `POST /api/portfolios/{id}/snapshots` → save current analytics summary as a JSON snapshot
- `GET /api/portfolios/{id}/snapshots` → list snapshots
- Ownership enforced via `PortfolioService.GetPortfolioSummary(userID, portfolioID)`
  - **403** if not owner
  - **404** if portfolio not found

### 4) DB migration

- `backend/migrations/schema.sql` updated to include `portfolio_snapshots` table (`IF NOT EXISTS`)

### 5) Server wiring

- `backend/cmd/server/main.go` wired repositories + handlers + routes
- Admin routes wrapped with `AuthMiddleware` + `RequireRole("ADMIN")`

### 6) Email reporting removed (per request)

- Removed SMTP mailer and email endpoint
- Removed route `POST /api/portfolios/{id}/report/email`

---

## Sprint 3 – Raghav’s Work (Frontend Admin)

### Overview
Implemented the frontend admin flow and test coverage for admin features.

### Key deliverables

- Built and integrated `AdminDashboardComponent` frontend flow for admin operations
- Added `/admin` route integration with existing auth/role-based guard behavior

#### User Management UI
- List users
- Update user role (`USER` / `ADMIN`)
- Success/error feedback messages

#### Asset Management UI
- Create asset form
- Edit existing asset
- Delete asset
- Refresh/reload behavior

#### Robust UI state handling
- Loading states
- Action-in-progress disabling
- Success/error toast/message handling

---

## Sprint 3 – Rishithaa Maligireddy’s Work (Frontend)

### Overview
Rishithaa’s Sprint 3 work focused on strengthening the **frontend experience** of WealthScope and integrating the **AI recommender system interface** into the website. Her contributions improved usability, visualization, responsiveness, and frontend readiness for intelligent portfolio insights.

### Key deliverables

1) **Reporting & Analytics dashboard**
- Display of total portfolio value
- Profit/loss visualization
- Percentage return metrics
- Asset allocation breakdown across stocks, bonds, and cash
- Export support for portfolio reports in PDF and CSV format

2) **UI/UX improvements**
- Implementation of a modern dark theme
- Improved layout and spacing
- Better navigation flow
- More responsive design across screen sizes and devices

3) **API integration + resilience**
- Improved frontend integration with backend APIs so that data could be fetched dynamically and reflected on the interface in real time
- Loading indicators during API requests
- User-friendly error messages when backend calls fail

4) **Modular frontend design**
- Structured using reusable components and services (improves maintainability and supports future expansion)

### AI Recommender System UI

#### Overview
Rishithaa also worked on the frontend layer for the AI recommender system, making it possible for users to interact with the recommendation flow directly from the website.

#### 1) Input Interface
The recommender interface allows users to:
- enter portfolio tickers
- choose a risk profile
- select an investment horizon
- trigger recommendation generation using a dedicated insights button

#### 2) Recommender Functionality
The frontend was designed to support recommendation display based on backend and ML outputs. The interface is intended to present:
- portfolio distribution insights
- asset allocation analysis
- performance-based observations
- diversification insights
- risk assessment
- optimization suggestions

#### 3) Workflow Integration
The AI recommender workflow on the frontend was structured as follows:
1. user enters portfolio-related inputs
2. frontend sends the data to the backend
3. backend communicates with the ML service
4. recommendation results are returned
5. frontend displays the generated insights clearly

#### 4) UI Behavior
- initially shows a **“No Insights Yet”** state
- after generation, displays recommendation insights in a structured format

### Outcomes of Rishithaa’s Sprint 3 Work
- Improved frontend usability and design quality
- Added a stronger analytics and reporting experience
- Successfully integrated the frontend interface for the AI recommender system
- Improved user interaction through better responsiveness and feedback states
- Strengthened frontend modularity and scalability

---

## Sprint 3 – Ansh Jain’s Work (AI Service Backend)

### Overview
My Sprint 3 work focused on upgrading the **WealthScope AI service backend** into a richer finance intelligence layer. The goal was to improve query understanding, retrieval quality, portfolio insight generation, grounded chatbot behavior, and structured AI-powered backend endpoints.

### Objectives
- Improve backend intelligence for finance-related queries
- Strengthen chatbot routing and response grounding
- Expand retrieval and knowledge integration
- Add portfolio insight and prediction features
- Build structured AI endpoints for frontend/dashboard use
- Improve backend reliability through modular design and testing

### 1. Entity Extraction Improvements
I implemented a stronger entity extraction flow for stock and company recognition. This included:
- ticker detection for `$AAPL`-style inputs
- plain ticker detection like `TSLA`
- company-name recognition such as `Apple -> AAPL`
- support for multiple mentions in one query
- false-positive prevention for common words

### 2. Improved Intent Routing with Fallback Support
I extended the intent classification pipeline so the AI service can use improved/remote intent routing while still falling back safely to the original keyword-based logic if the remote/model path fails. This made routing more robust for chat, market-data enrichment, and retrieval decisions.

### 3. Retrieval and RAG Upgrades
I expanded the retrieval layer by adding modular retrieval components for:
- chunking
- lexical retrieval
- hybrid retrieval
- TF-IDF indexing
- backward-compatible retrieval behavior

I also integrated the upgraded retrieval flow into the chat pipeline so answers could be better grounded in relevant financial knowledge.

### 4. Portfolio Risk Drift Prediction
I added backend support for **portfolio risk drift prediction**, allowing the AI service to estimate whether a portfolio is drifting away from its intended risk profile and return structured explanatory output.

### 5. Multi-Stock Comparison Feature
I added backend support for comparing multiple stocks in a structured way. This allows the system to compare symbols across market and company-level characteristics and return a neutral summary suitable for frontend use and chatbot grounding.

### 6. Portfolio Explanation Endpoint
I implemented a portfolio explanation layer so the backend can generate structured, human-readable portfolio insights such as:
- concentration risk
- diversification notes
- risk alignment commentary
- neutral, non-advisory explanations

### 7. Source-Grounded Chat Response Formatting
I improved the grounded response construction used by the chatbot by validating prompt section order, ensuring section headers appear correctly, and supporting structured knowledge/context injection. This helps the chatbot generate more reliable, finance-scoped responses.

### 8. News Sentiment Aggregation
I added ticker-level news sentiment support so the backend can aggregate article-level sentiment and expose structured sentiment information for a stock/news workflow.

### 9. Session and Chat Memory Improvements
I improved the backend chat session behavior by strengthening:
- session creation
- session reuse
- session clearing
- session expiration
- compaction/summarization of older session history

### 10. CSV-Backed Finance Knowledge Base Integration
I integrated support for a CSV-backed finance Q&A knowledge base into retrieval and chat grounding. The backend now includes loading and retrieval validation for the QA dataset and can inject relevant QA knowledge into the chat flow.

### 11. Dashboard-Facing AI Backend Support
I added and validated backend support for structured AI endpoints beyond the general chat route, including:
- comparison
- portfolio explanation
- portfolio summarization
- portfolio change analysis
- news sentiment
- risk drift-related functionality

These endpoints are designed to support frontend dashboard integration.

### 12. Frontend Chatbot Icon and Chatbot UI Integration
I also added frontend-facing chatbot integration planning and implementation support by introducing a chatbot entry point through a dedicated floating chatbot icon and chatbot UI concept. This involved:
- adding a floating chatbot icon/button on the website
- using a custom chatbot character/icon named **Leo**
- opening a chatbot panel or modal when the icon is clicked
- designing the chatbot interface with:
  - header/title
  - message area
  - user and bot message bubbles
  - input field
  - send button
  - close/minimize behavior
- connecting the frontend chatbot UI to the backend `/chat` API

This improves accessibility and usability by making the AI assistant directly available from the frontend dashboard experience.

### Outcomes of My Sprint 3 Work
- Significantly improved AI-service intelligence and routing quality
- Stronger entity extraction and intent classification behavior
- Better retrieval-backed and grounded chatbot responses
- Added practical portfolio analytics features such as explanation and risk drift prediction
- Expanded structured backend AI APIs for frontend integration
- Strengthened backend reliability with modular code changes and extensive test coverage

---

## List Frontend Unit Tests

### Rishithaa Maligireddy
Frontend unit tests for dashboard and recommender UI components should include:
- analytics dashboard component rendering
- loading state rendering
- error state rendering
- AI recommender input form validation
- “No Insights Yet” state behavior
- recommendation result rendering
- responsive layout/component behavior
- API service call mocking for frontend components

### Ansh Jain
Frontend chatbot UI integration was added as part of Sprint 3 scope, including the chatbot icon/button and chatbot panel design/integration. However, no separate frontend unit tests were added yet as part of this work, since verified testing remained focused on the backend AI service.

---

## List Backend Unit Tests

### Backend unit tests for Ansh Jain’s work included

#### Chat prompt / grounding tests
- `TestBuildUserContent_SectionOrderAndHeaders`
- `TestBuildUserContent_MissingKnowledge`
- `TestBuildUserContent_SectionHeadersAppearOnce`
- `TestBuildUserContent_MissingMarketAndNews`

#### Compare handler / comparison tests
- `TestHTTPStatusForCompareError`
- `TestNormalizeAndValidate_TwoSymbols`
- `TestNormalizeAndValidate_FourSymbols`
- `TestNormalizeAndValidate_TooFew`
- `TestNormalizeAndValidate_TooMany`
- `TestNormalizeAndValidate_EmptyString`
- `TestNormalizeAndValidate_DuplicatesCollapseToOne`
- `TestNormalizeAndValidate_MissingSymbolsNil`
- `TestCompare_MockSuccess`
- `TestCompare_QuoteFailure`
- `TestCompare_SummaryNoBuySellLanguage`
- `TestHTTP_Compare_InvalidCount`
- `TestHTTP_Compare_OK_WithMockFetcher`
- `TestHTTP_Compare_BadJSON`
- `TestHTTP_Compare_UpstreamQuoteError_Is502`

#### OpenAI / session store tests
- `TestCallOpenAI_Success`
- `TestCallOpenAI_NonOKStatus`
- `TestCallOpenAI_EmptyChoices`
- `TestCallOpenAI_InvalidJSONBody`
- `TestStore_SessionCreation`
- `TestStore_SessionReuse`
- `TestStore_ClearSession`
- `TestStore_SessionExpiration`
- `TestStore_CompactionSummarizesOldest`
- `TestFoldOldestIntoSummary_Shape`
- `TestDefaultStore_ClearIntegration`
- `TestGetSystemPrompt_MentionsGroundingSections`

#### Portfolio summary / changes tests
- `TestDescribeChanges_NoPrior`
- `TestDescribeChanges_WithPrior`
- `TestDescribeChanges_EmptyCurrent`
- `TestSummarize_Basic`
- `TestSummarize_EmptyHoldings`

#### CSV QA dataset / retrieval tests
- `TestLoadQADatasetFromPath_InvalidHeader`
- `TestLoadQADatasetFromPath_OK`
- `TestRetrieveQAWithContext_FromTempFile`
- `TestLoadQADatasetFromPath_EmptyDataRows`
- `TestLoadQADatasetFromPath_WrongFieldCount`
- `TestFormatQAKnowledgeLine_Truncates`
- `TestRetrieveQA_RealDatasetIfPresent`
- `TestBuildUserContent_QASectionWithRetrievalStyleLine`

#### Chat / service integration tests
- `TestBuildEnvelopeInputForChat_IncludesKnowledgeAndQA`
- `TestBuildEnvelopeInputForChat_StockPriceIntentUsesStubbedMarket`
- `TestBuildEnvelopeInputForChat_NoMarketEnrichmentForGenericQuestion`

#### Chat handler tests
- `TestChatHandler_Success`
- `TestChatHandler_EmptyMessage`
- `TestChatHandler_BadJSON`
- `TestChatHandler_ServiceError`
- `TestClearChatHandler_Success`
- `TestChatHandler_DefaultSession`

#### Entity extraction tests
- `TestExtract_DollarTicker`
- `TestExtract_PlainTSLA`
- `TestExtract_CompanyNameApple`
- `TestExtract_MultipleAppleAndMicrosoft`
- `TestExtract_NoTicker`
- `TestExtract_FalsePositivePrevention`

#### Intent routing tests
- `TestDetectIntent_RemotePrimary`
- `TestDetectIntent_RemoteFallbackOnError`
- `TestDetectIntent_RemoteFallbackInvalidJSON`
- `TestDetectIntent_RemoteFallbackUnknownLabel`
- `TestDetectIntent_StockPrice`
- `TestDetectIntent_RiskAnalysis`
- `TestDetectIntent_MarketNews`
- `TestDetectIntent_PortfolioTip`
- `TestDetectIntent_GeneralMarket`
- `TestDetectIntent_Unknown`
- `TestDetectIntent_TickerExtracted`
- `TestDetectIntent_ConfidenceNonZero`
- `TestDetectIntent_NoRemoteSameAsKeywords`

#### News sentiment tests
- `TestNewsSentiment_AggregateBullish`
- `TestNewsSentiment_AggregateBearish`
- `TestNewsSentiment_AggregateMixedBullAndBearPopulatesBothTops`
- `TestNewsSentiment_AggregateNeutral`
- `TestNewsSentiment_EmptyArticles`
- `TestLexicalSentimentScores_Exported`
- `TestHTTP_NewsSentiment_BadGatewayOrOK`
- `TestHTTP_NewsSentiment_OK_WithMockFetcher`
- `TestHTTP_NewsSentiment_FetchError_Is502`

#### Portfolio explanation tests
- `TestPortfolioExplain_HighRiskConcentrated`
- `TestPortfolioExplain_Balanced`
- `TestPortfolioExplain_InvalidEmptyHoldings`
- `TestPortfolioExplain_InvalidTarget`
- `TestHTTP_PortfolioExplain_OK`
- `TestHTTP_PortfolioExplain_BadJSON`
- `TestHTTP_PortfolioExplain_EmptyHoldings`

#### RAG / retrieval tests
- `TestRetrieveWithContext_SemanticBetaQuery`
- `TestRetrieveWithContext_EntityBoostTicker`
- `TestRetrieve_BackwardCompatible`
- `TestRetrieve_ReturnsResults`
- `TestRetrieve_TopKRespected`
- `TestRetrieve_RelevantTopicReturned`
- `TestRetrieve_NoMatchReturnsEmpty`
- `TestRetrieve_PERatioQuery`

#### Risk drift prediction tests
- `TestPredictRiskDrift_LowTargetHighBeta`
- `TestPredictRiskDrift_AlignedMedium`
- `TestPredictRiskDrift_NormalizesWeights`
- `TestPredictRiskDrift_InvalidTarget`
- `TestHTTP_RiskDrift_OK`
- `TestHTTP_RiskDrift_EmptyHoldings`
- `TestHTTP_RiskDrift_BadJSON`
- `TestHTTP_RiskDrift_InvalidTarget`

#### Existing backend logic tests still passing after integration
- `TestRisk_HighRisk`
- `TestRisk_MediumRisk`
- `TestRisk_LowRisk`
- `TestRisk_InvalidBetaDefaultsToOne`
- `TestRisk_ExplanationNotEmpty`
- `TestRisk_ScoreCalculation`
- `TestSentiment_Bullish`
- `TestSentiment_Bearish`
- `TestSentiment_Neutral`
- `TestSentiment_BullishOverBearish`
- `TestSentiment_EmptyText`
- `TestExtractTicker_DollarFormat`
- `TestExtractTicker_PlainFormat`
- `TestExtractTicker_NoTicker`
- `TestExtractTicker_IgnoresCommonWords`
- `TestExtractTicker_DollarTakesPriority`

---

## Updated Documentation for Backend API (AI Service)

### Base URL
`http://localhost:9000`

### Existing and Updated AI-Service Endpoints
- `GET /health` – service health check
- `POST /chat` – chatbot interaction endpoint
- `DELETE /chat/session/:session_id` – clear chat session
- `POST /intent` – detect finance-related intent
- `POST /sentiment` – sentiment analysis
- `POST /risk` – portfolio risk scoring
- `POST /predict/risk-drift` – risk drift prediction
- `GET /quote/:symbol` – stock quote
- `GET /company/:symbol` – company overview
- `GET /news/:symbol` – market news
- `POST /compare` – multi-stock comparison
- `POST /portfolio/explain` – portfolio explanation
- `POST /portfolio/summarize` – concise portfolio summary
- `POST /portfolio/changes` – describe portfolio changes
- `GET /news-sentiment/:symbol` – aggregate news sentiment

### Frontend Integration Note
The frontend now includes:
- analytics dashboard and recommender UI integration through Rishithaa’s work
- chatbot entry-point integration through a floating chatbot icon and chatbot panel UI through Ansh’s work

Together, these changes improve both the frontend user experience and the backend AI intelligence capabilities of WealthScope.

---

## Overall Sprint 3 Outcomes
- Stronger and more scalable frontend experience
- Improved analytics and recommender usability
- Significantly upgraded backend AI-service intelligence
- Better routing, retrieval, grounding, and portfolio insight generation
- Expanded product-facing AI capabilities across both backend and frontend

---

## Future Work
- Real-time data integration
- Improved ML model accuracy
- Enhanced visualization of AI insights
- Personalized recommendation history
- Stronger frontend testing for chatbot and recommender flows

---

## Tests / Verification (Sprint 3)

### Backend (Go) – Vittesh

Run unit tests:

```bash
cd backend
gofmt -w .
go test ./...
```

Admin route existence (expect 401 without token):

```bash
BASE="http://localhost:8080/api"
curl -i -s "$BASE/admin/users"
```

Login and verify admin access:

```bash
BASE="http://localhost:8080/api"

ACCESS_TOKEN=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"vittesharora@test.com","password":"test123"}' | jq -r '.access_token')

curl -i -s "$BASE/admin/users" -H "Authorization: Bearer $ACCESS_TOKEN"
```

Snapshot create/list:

```bash
PORTFOLIO_ID="<portfolio-id>"

curl -i -s -X POST "$BASE/portfolios/$PORTFOLIO_ID/snapshots" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl -i -s "$BASE/portfolios/$PORTFOLIO_ID/snapshots" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Frontend unit tests – Raghav (Admin)

- `src/app/services/admin.service.spec.ts`
  - `getUsers` maps backend rows to `AdminUser`
  - `updateUserRole` PATCHes role endpoint
  - `getAssets` maps backend rows to `AdminAsset`
  - `createAsset` POSTs mapped payload and maps response
  - `updateAsset` PUTs mapped payload
  - `deleteAsset` DELETEs asset by id

- `src/app/features/admin-dashboard.component.spec.ts`
  - loads users/assets on init
  - `reloadAll` clears messages and reloads data
  - `saveUserRole` sets success message on success
  - `saveUserRole` sets error message on failure
  - `submitAssetForm` creates when not editing
  - `submitAssetForm` updates when `editAssetId` is set
  - `startEditAsset` populates the form
  - `removeAsset` refreshes assets on success
  - `resetAssetForm` clears edit state and form

### Cypress E2E – Raghav (Admin)

- `cypress/e2e/admin-dashboard.cy.js`
  - shows the dashboard
  - switches to the Assets tab

---

## Updated API Summary (for this repo backend)

### Base URL
`http://localhost:8080/api`

### Admin (ADMIN-only)
- `GET /admin/users`
- `PATCH /admin/users/{id}/role`
- `GET /admin/assets`
- `POST /admin/assets`
- `PUT /admin/assets/{id}`
- `DELETE /admin/assets/{id}`

### Portfolio snapshots (authenticated, owner-only)
- `POST /portfolios/{id}/snapshots`
- `GET /portfolios/{id}/snapshots`


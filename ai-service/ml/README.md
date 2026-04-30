# Intent classifier (TF-IDF + Logistic Regression)

Trained intent model that classifies chat messages into one of:

`STOCK_PRICE`, `RISK_ANALYSIS`, `MARKET_NEWS`, `PORTFOLIO_TIP`, `GENERAL_MARKET`, `UNKNOWN`

The Go AI service (`internal/ml/intent.go`) calls this Python service first;
keyword fallback runs if the HTTP call fails or confidence is below threshold.

## Layout

| File | Purpose |
|------|---------|
| `intent_dataset.csv` | Labeled examples; header `text,label` |
| `train_intent_model.py` | TF-IDF (1-2 grams) + Logistic Regression trainer |
| `intent_inference_service.py` | Flask app exposing `POST /classify-intent` and `GET /health` |
| `intent_vectorizer.joblib` / `intent_model.joblib` / `intent_labels.json` | Trained artifacts |
| `test_pipeline.py` | Smoke tests for training + inference |
| `requirements.txt` | Python dependencies |

## Setup

```bash
cd ai-service/ml
python -m venv .venv
# Windows PowerShell
.\.venv\Scripts\Activate.ps1
# macOS / Linux
source .venv/bin/activate

pip install -r requirements.txt
```

## Train

```bash
python train_intent_model.py
```

Optional flags:

```bash
python train_intent_model.py --data ./intent_dataset.csv --out-dir .
```

The script prints accuracy, macro-F1, classification report, and confusion matrix.
Artifacts are written next to the script.

## Serve

```bash
# default port 8088
python intent_inference_service.py

# custom port
PORT=9001 python intent_inference_service.py
```

### Endpoints

`POST /classify-intent`

```bash
curl -s -X POST http://localhost:8088/classify-intent \
  -H 'Content-Type: application/json' \
  -d '{"message":"What is AAPL trading at?"}'
# {"intent":"STOCK_PRICE","confidence":0.83}
```

`GET /health`

```bash
curl -s http://localhost:8088/health
# {"status":"ok","labels":[...],"model_classes":[...]}
```

Errors:
- `400` — empty/whitespace `message`
- `500` — model returned a label outside the whitelist

## Wiring into the Go service

The AI service reads two env vars:

| Var | Purpose | Example |
|-----|---------|---------|
| `INTENT_CLASSIFIER_URL` | Base URL of this Flask service. Empty ⇒ keyword-only. | `http://localhost:8088` |
| `INTENT_MIN_CONFIDENCE` | Float in `[0,1]`. Remote predictions below this fall back to keyword scorer. `0` (default) disables. | `0.45` |

```bash
# from ai-service/
$env:INTENT_CLASSIFIER_URL = "http://localhost:8088"
$env:INTENT_MIN_CONFIDENCE = "0.45"
go run ./cmd
```

## Tests

```bash
# Python — train + inference smoke
python test_pipeline.py

# Go — keyword + remote fallback paths
cd ..
go test ./tests/...
```

## Retraining workflow

1. Edit `intent_dataset.csv` (keep header `text,label`; only the 6 labels).
2. `python train_intent_model.py` — overwrites the `.joblib` artifacts.
3. Restart the Flask service so it loads the new artifacts.
4. Re-run `go test ./tests/...` to confirm the contract still holds.

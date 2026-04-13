"""
Lightweight HTTP service: POST /classify-intent
"""
from __future__ import annotations

import json
from pathlib import Path

import joblib
from flask import Flask, jsonify, request

ML_DIR = Path(__file__).resolve().parent


def load_artifacts():
    labels_path = ML_DIR / "intent_labels.json"
    with labels_path.open(encoding="utf-8") as f:
        data = json.load(f)
    allowed = set(data.get("labels", []))
    vectorizer = joblib.load(ML_DIR / "intent_vectorizer.joblib")
    model = joblib.load(ML_DIR / "intent_model.joblib")
    return allowed, vectorizer, model


ALLOWED_LABELS, VECTORIZER, MODEL = load_artifacts()

app = Flask(__name__)


@app.post("/classify-intent")
def classify_intent():
    payload = request.get_json(silent=True) or {}
    message = (payload.get("message") or "").strip()
    if not message:
        return jsonify({"error": "message required"}), 400

    x = VECTORIZER.transform([message])
    pred = MODEL.predict(x)[0]
    if pred not in ALLOWED_LABELS:
        return jsonify({"error": "invalid prediction"}), 500

    proba = MODEL.predict_proba(x)[0]
    classes = list(MODEL.classes_)
    idx = classes.index(pred)
    confidence = float(min(1.0, max(0.0, proba[idx])))

    return jsonify({"intent": pred, "confidence": confidence})


def main():
    import os

    port = int(os.environ.get("PORT", "8088"))
    app.run(host="0.0.0.0", port=port, debug=False)


if __name__ == "__main__":
    main()

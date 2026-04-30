"""
FINAL Optimized Training Script
- Removes duplicates
- Uses Stratified Cross Validation
- Uses GridSearchCV with f1_macro
- No train/test leakage
"""

from __future__ import annotations

import argparse
import json
from pathlib import Path

import pandas as pd
import joblib

from sklearn.pipeline import Pipeline
from sklearn.model_selection import GridSearchCV, StratifiedKFold
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression


LABEL_ORDER = [
    "STOCK_PRICE",
    "RISK_ANALYSIS",
    "MARKET_NEWS",
    "PORTFOLIO_TIP",
    "GENERAL_MARKET",
    "UNKNOWN",
]


def train_and_save(data_path: Path, out_dir: Path):

    # ======================
    # 1. LOAD + CLEAN DATA
    # ======================
    df = pd.read_csv(data_path)

    # 🔥 remove duplicates
    df = df.drop_duplicates(subset=["text"])

    texts = df["text"].str.lower().tolist()
    labels = df["label"].tolist()

    print(f"Dataset size after cleaning: {len(texts)}")

    # ======================
    # 2. PIPELINE
    # ======================
    pipeline = Pipeline([
        ("tfidf", TfidfVectorizer()),
        ("clf", LogisticRegression(max_iter=2000))
    ])

    # ======================
    # 3. PARAM GRID
    # ======================
    param_grid = {
        "tfidf__ngram_range": [(1, 1), (1, 2)],
        "tfidf__min_df": [1, 2],
        "tfidf__max_df": [0.85, 0.95, 1.0],
        "tfidf__max_features": [5000, 8000],

        "clf__C": [0.5, 1, 2, 5],
        "clf__solver": ["lbfgs"],
        "clf__class_weight": ["balanced"]
    }

    # ======================
    # 4. STRATIFIED CV
    # ======================
    cv = StratifiedKFold(n_splits=5, shuffle=True, random_state=42)

    grid = GridSearchCV(
        pipeline,
        param_grid,
        cv=cv,
        n_jobs=-1,
        verbose=2,
        scoring="f1_macro"
    )

    # ======================
    # 5. TRAIN
    # ======================
    grid.fit(texts, labels)

    best_model = grid.best_estimator_

    print("\n🔥 Best Params:", grid.best_params_)
    print("🔥 Best CV Score (Macro F1):", grid.best_score_)

    # ======================
    # 6. SAVE MODEL
    # ======================
    out_dir.mkdir(parents=True, exist_ok=True)

    joblib.dump(best_model.named_steps["tfidf"], out_dir / "intent_vectorizer.joblib")
    joblib.dump(best_model.named_steps["clf"], out_dir / "intent_model.joblib")

    with (out_dir / "intent_labels.json").open("w") as f:
        json.dump({"labels": LABEL_ORDER}, f, indent=2)

    print("\n✅ FINAL optimized model saved!")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--data", type=Path, default="augmented_dataset.csv")
    parser.add_argument("--out-dir", type=Path, default=".")
    args = parser.parse_args()

    train_and_save(args.data, args.out_dir)


if __name__ == "__main__":
    main()

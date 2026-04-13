"""
Train TF-IDF + Logistic Regression intent classifier and save sklearn artifacts.
"""
from __future__ import annotations

import argparse
import csv
import json
from pathlib import Path

import joblib
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import accuracy_score, classification_report, confusion_matrix, f1_score
from sklearn.model_selection import train_test_split

LABEL_ORDER = [
    "STOCK_PRICE",
    "RISK_ANALYSIS",
    "MARKET_NEWS",
    "PORTFOLIO_TIP",
    "GENERAL_MARKET",
    "UNKNOWN",
]


def load_rows(csv_path: Path) -> tuple[list[str], list[str]]:
    texts: list[str] = []
    labels: list[str] = []
    with csv_path.open(newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        fields = [h.strip() for h in (reader.fieldnames or [])]
        if fields != ["text", "label"]:
            raise ValueError(f"Expected header text,label got {reader.fieldnames}")
        for row in reader:
            t = (row.get("text") or "").strip()
            lab = (row.get("label") or "").strip()
            if not t or not lab:
                continue
            if lab not in LABEL_ORDER:
                raise ValueError(f"Unknown label {lab!r}")
            texts.append(t)
            labels.append(lab)
    if len(texts) < 10:
        raise ValueError("Need at least 10 labeled rows")
    return texts, labels


def train_and_save(data_path: Path, out_dir: Path, *, verbose: bool = True) -> None:
    texts, y = load_rows(data_path)
    x_train, x_test, y_train, y_test = train_test_split(
        texts,
        y,
        test_size=0.2,
        random_state=42,
        stratify=y,
    )

    vectorizer = TfidfVectorizer(
        ngram_range=(1, 2),
        min_df=1,
        max_features=8000,
        sublinear_tf=True,
    )
    x_train_v = vectorizer.fit_transform(x_train)
    x_test_v = vectorizer.transform(x_test)

    model = LogisticRegression(
        max_iter=2000,
        class_weight="balanced",
        random_state=42,
    )
    model.fit(x_train_v, y_train)
    pred = model.predict(x_test_v)

    if verbose:
        acc = accuracy_score(y_test, pred)
        macro_f1 = f1_score(y_test, pred, average="macro", labels=LABEL_ORDER, zero_division=0)
        print(f"Accuracy: {acc:.4f}")
        print(f"Macro F1: {macro_f1:.4f}")
        print("\nClassification report:")
        print(classification_report(y_test, pred, labels=LABEL_ORDER, zero_division=0))
        cm = confusion_matrix(y_test, pred, labels=LABEL_ORDER)
        print("\nConfusion matrix (rows=true, cols=pred):")
        print(cm)

    out_dir.mkdir(parents=True, exist_ok=True)
    joblib.dump(vectorizer, out_dir / "intent_vectorizer.joblib")
    joblib.dump(model, out_dir / "intent_model.joblib")
    with (out_dir / "intent_labels.json").open("w", encoding="utf-8") as f:
        json.dump({"labels": LABEL_ORDER}, f, indent=2)
    if verbose:
        print(f"\nWrote artifacts to {out_dir}")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--data",
        type=Path,
        default=Path(__file__).resolve().parent / "intent_dataset.csv",
    )
    parser.add_argument(
        "--out-dir",
        type=Path,
        default=Path(__file__).resolve().parent,
    )
    args = parser.parse_args()
    train_and_save(args.data, args.out_dir, verbose=True)


if __name__ == "__main__":
    main()

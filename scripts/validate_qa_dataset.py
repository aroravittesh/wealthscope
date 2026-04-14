#!/usr/bin/env python3
"""Validate data/qa_dataset.csv; optional --sample for random QA review."""
from __future__ import annotations

import argparse
import csv
import random
import sys
from collections import Counter
from pathlib import Path

HEADER = [
    "id",
    "category",
    "sub_category",
    "question",
    "answer",
    "keywords",
    "ticker",
    "difficulty",
    "source_type",
    "priority",
    "last_updated",
]
ALLOW_DIFF = {"beginner", "intermediate", "advanced"}
ALLOW_ST = {
    "educational",
    "risk_guidance",
    "finance_concept",
    "market_term",
    "platform_help",
    "compliance",
}
ALLOW_PRI = {"high", "medium", "low"}
EXPECTED_DATE = "2026-04-12"
EXPECTED_ROWS = 800
EXPECTED_CATS = 20
EXPECTED_PER_CAT = 40

FINANCE_TEMPLATE = " is a standard market concept worth understanding for context."


def load_rows(path: Path) -> tuple[list[str], list[list[str]]]:
    with path.open(encoding="utf-8", newline="") as f:
        reader = csv.reader(f)
        header = next(reader)
        rows = list(reader)
    return header, rows


def validate(path: Path) -> list[tuple[str, str]]:
    issues: list[tuple[str, str]] = []
    header, rows = load_rows(path)

    if header != HEADER:
        issues.append(("header", f"expected {HEADER}, got {header}"))

    if len(rows) != EXPECTED_ROWS:
        issues.append(("row_count", f"expected {EXPECTED_ROWS}, got {len(rows)}"))

    exp_ids = [f"QA{i:04d}" for i in range(1, EXPECTED_ROWS + 1)]
    ids = [r[0] for r in rows if len(r) > 0]
    if ids != exp_ids:
        first = next(((a, b) for a, b in zip(ids, exp_ids) if a != b), None)
        issues.append(("id_sequence", f"first mismatch {first}; got {len(ids)} ids"))

    if len(set(ids)) != len(ids):
        issues.append(("id_unique", "duplicate ids present"))

    for i, r in enumerate(rows, start=2):
        if len(r) != 11:
            issues.append(("column_count", f"line {i}: expected 11 fields, got {len(r)}"))
            break

    seen_q: set[str] = set()
    for i, r in enumerate(rows, start=2):
        if len(r) < 11:
            continue
        q, a = r[3], r[4]
        if not str(q).strip():
            issues.append(("empty_question", f"row line {i} id={r[0]}"))
        if not str(a).strip():
            issues.append(("empty_answer", f"row line {i} id={r[0]}"))
        if q in seen_q:
            issues.append(("duplicate_question", f"line {i} id={r[0]}"))
        seen_q.add(q)

    for i, r in enumerate(rows, start=2):
        if len(r) < 11:
            continue
        diff, st, pri, lu = r[7], r[8], r[9], r[10]
        rid = r[0]
        if diff not in ALLOW_DIFF:
            issues.append(("difficulty", f"{rid}: {diff!r}"))
        if st not in ALLOW_ST:
            issues.append(("source_type", f"{rid}: {st!r}"))
        if pri not in ALLOW_PRI:
            issues.append(("priority", f"{rid}: {pri!r}"))
        if lu != EXPECTED_DATE:
            issues.append(("last_updated", f"{rid}: {lu!r}"))

    cc = Counter(r[1] for r in rows if len(r) > 1)
    if len(cc) != EXPECTED_CATS:
        issues.append(("category_count", f"expected {EXPECTED_CATS} categories, got {len(cc)}"))
    for cat, n in sorted(cc.items()):
        if n != EXPECTED_PER_CAT:
            issues.append(("category_balance", f"{cat!r}: {n} (expected {EXPECTED_PER_CAT})"))

    return issues


def sample_review(rows: list[list[str]], seed: int, n: int) -> None:
    random.seed(seed)
    idxs = sorted(random.sample(range(len(rows)), min(n, len(rows))))
    print(f"\n=== {len(idxs)} random rows (seed={seed}) ===\n")
    for i in idxs:
        r = rows[i]
        rid, cat, sub, q, a = r[0], r[1], r[2], r[3], r[4]
        wc = len(a.split())
        flags: list[str] = []
        if FINANCE_TEMPLATE in a:
            flags.append("template_answer")
        if any(
            s in q
            for s in (
                " when my focus is ",
                " when I'm reviewing ",
                " when the context is ",
                " when reviewing ",
            )
        ):
            flags.append("keyword_stuffed_question")
        if wc < 50:
            flags.append("short_answer")
        flag_str = ", ".join(flags) if flags else "none"
        print(f"{rid} | {cat} | {sub}")
        print(f"  flags: {flag_str}")
        print(f"  Q: {q[:280]}{'...' if len(q) > 280 else ''}")
        print(f"  A ({wc} words): {a[:240].replace(chr(10), ' ')}{'...' if len(a) > 240 else ''}")
        print()


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--csv", type=Path, default=Path("data/qa_dataset.csv"))
    ap.add_argument("--sample", type=int, default=0, help="random rows to print for review")
    ap.add_argument("--seed", type=int, default=42)
    args = ap.parse_args()
    root = Path(__file__).resolve().parents[1]
    path = args.csv if args.csv.is_absolute() else root / args.csv

    issues = validate(path)
    if issues:
        print(f"FAIL: {len(issues)} issue(s)")
        for code, msg in issues[:50]:
            print(f"  [{code}] {msg}")
        if len(issues) > 50:
            print(f"  ... and {len(issues) - 50} more")
        return 1

    print(f"OK: {path} — {EXPECTED_ROWS} rows, header, ids, enums, balance.")

    _, rows = load_rows(path)
    ans_c = Counter(r[4] for r in rows)
    dup_ans = sum(1 for c in ans_c.values() if c > 1)
    tmpl = sum(1 for r in rows if FINANCE_TEMPLATE in r[4])
    print(f"Stats: duplicate identical answers={dup_ans}, finance template answers={tmpl}")

    if args.sample > 0:
        sample_review(rows, args.seed, args.sample)

    return 0


if __name__ == "__main__":
    sys.exit(main())

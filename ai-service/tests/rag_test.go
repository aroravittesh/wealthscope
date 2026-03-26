package tests

import (
    "testing"
    "wealthscope-ai/internal/rag"
)

func TestRetrieve_ReturnsResults(t *testing.T) {
    results := rag.Retrieve("What is beta and volatility?", 3)
    if len(results) == 0 {
        t.Fatal("expected at least one result")
    }
}

func TestRetrieve_TopKRespected(t *testing.T) {
    results := rag.Retrieve("stock market portfolio dividend", 2)
    if len(results) > 2 {
        t.Fatalf("expected max 2 results got %d", len(results))
    }
}

func TestRetrieve_RelevantTopicReturned(t *testing.T) {
    results := rag.Retrieve("dividend yield income", 3)
    found := false
    for _, r := range results {
        if r.Topic == "dividends" {
            found = true
        }
    }
    if !found {
        t.Fatal("expected dividends topic to be returned")
    }
}

func TestRetrieve_NoMatchReturnsEmpty(t *testing.T) {
    results := rag.Retrieve("xyzzy random nonsense", 3)
    if len(results) != 0 {
        t.Fatalf("expected 0 results got %d", len(results))
    }
}

func TestRetrieve_PERatioQuery(t *testing.T) {
    results := rag.Retrieve("what is PE ratio valuation", 3)
    found := false
    for _, r := range results {
        if r.Topic == "pe_ratio" {
            found = true
        }
    }
    if !found {
        t.Fatal("expected pe_ratio topic to be returned")
    }
}
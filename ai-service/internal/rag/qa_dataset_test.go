package rag

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadQADatasetFromPath_InvalidHeader(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.csv")
	if err := os.WriteFile(p, []byte("a,b,c\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadQADatasetFromPath(p)
	if err == nil {
		t.Fatal("expected error for bad header")
	}
}

func TestLoadQADatasetFromPath_OK(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "qa.csv")
	csv := strings.Join([]string{
		strings.Join(qaHeader, ","),
		`QA0999,Cat One,sub_a,"What is X?","X is a concept.","kw1,kw2",,beginner,educational,high,2026-04-12`,
	}, "\n")
	if err := os.WriteFile(p, []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	chunks, err := LoadQADatasetFromPath(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 1 {
		t.Fatalf("want 1 chunk got %d", len(chunks))
	}
	ch := chunks[0]
	if ch.ID != "QA0999" {
		t.Fatalf("id %q", ch.ID)
	}
	if ch.Topic != "Cat One / sub_a" {
		t.Fatalf("topic %q", ch.Topic)
	}
	q, a := ChunkQAPair(ch)
	if q != "What is X?" || !strings.Contains(a, "X is a concept") {
		t.Fatalf("pair q=%q a=%q", q, a)
	}
	found := false
	for _, tag := range ch.Tags {
		if tag == "kw1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("tags %#v", ch.Tags)
	}
}

func TestRetrieveQAWithContext_FromTempFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "qa.csv")
	csv := strings.Join([]string{
		strings.Join(qaHeader, ","),
		`QA1000,Diversification,corr,"What is diversification?","Spreading investments reduces concentration risk.","diversification,risk",,beginner,educational,high,2026-04-12`,
		`QA1001,Other,other,"What is the weather?","Sunny sometimes.","weather",,beginner,educational,low,2026-04-12`,
	}, "\n")
	if err := os.WriteFile(p, []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}

	SetQADatasetPathForTest(p)
	t.Cleanup(ClearQADatasetPathOverride)

	hits := RetrieveQAWithContext("explain diversification and portfolio risk", RetrievalContext{}, 2)
	if len(hits) == 0 {
		t.Fatal("expected QA hits")
	}
	if hits[0].ID != "QA1000" {
		t.Fatalf("expected QA1000 first, got %s", hits[0].ID)
	}
}

func TestLoadQADatasetFromPath_EmptyDataRows(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "empty.csv")
	body := strings.Join(qaHeader, ",") + "\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadQADatasetFromPath(p)
	if err == nil {
		t.Fatal("expected error for zero data rows")
	}
}

func TestLoadQADatasetFromPath_WrongFieldCount(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "badrow.csv")
	lines := []string{
		strings.Join(qaHeader, ","),
		"only,one",
	}
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadQADatasetFromPath(p)
	if err == nil {
		t.Fatal("expected error for malformed row width")
	}
}

func TestFormatQAKnowledgeLine_Truncates(t *testing.T) {
	ch := KnowledgeChunk{ID: "QA0001", Topic: "A / B"}
	q := strings.Repeat("w", 300)
	a := strings.Repeat("x", 500)
	line := FormatQAKnowledgeLine(ch, q, a)
	if !strings.Contains(line, "QA0001") || !strings.Contains(line, "...") {
		snip := line
		if len(snip) > 120 {
			snip = snip[:120]
		}
		t.Fatalf("expected id and ellipsis truncation: %s", snip)
	}
}

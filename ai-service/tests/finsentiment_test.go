package tests

import (
	"testing"

	"wealthscope-ai/internal/finsentiment"
	"wealthscope-ai/internal/market"
)

func TestScoreText_BullishPhraseDominates(t *testing.T) {
	s := finsentiment.ScoreText("Apple beat estimates and raised guidance for the year")
	if s.Polarity <= 0 {
		t.Fatalf("expected positive polarity, got %f", s.Polarity)
	}
	if s.Bucket() != finsentiment.Bullish {
		t.Fatalf("expected BULLISH, got %s", s.Bucket())
	}
	if s.Bullish == 0 {
		t.Fatalf("expected weighted bullish hits > 0")
	}
}

func TestScoreText_BearishPhraseDominates(t *testing.T) {
	s := finsentiment.ScoreText("Company missed estimates and cut guidance amid layoffs announced")
	if s.Polarity >= 0 {
		t.Fatalf("expected negative polarity, got %f", s.Polarity)
	}
	if s.Bucket() != finsentiment.Bearish {
		t.Fatalf("expected BEARISH, got %s", s.Bucket())
	}
}

func TestScoreText_NeutralReturnsZero(t *testing.T) {
	s := finsentiment.ScoreText("The company released its quarterly report")
	if s.Polarity != 0 {
		t.Fatalf("expected polarity 0, got %f", s.Polarity)
	}
	if s.Bucket() != finsentiment.Neutral {
		t.Fatalf("expected NEUTRAL, got %s", s.Bucket())
	}
}

func TestScoreText_EmptyReturnsZero(t *testing.T) {
	s := finsentiment.ScoreText("")
	if s.Polarity != 0 || s.Bullish != 0 || s.Bearish != 0 {
		t.Fatalf("empty input must produce zero score, got %+v", s)
	}
}

func TestScoreText_NegationFlipsBullish(t *testing.T) {
	plain := finsentiment.ScoreText("growth")
	negated := finsentiment.ScoreText("no growth")
	if plain.Polarity <= 0 {
		t.Fatalf("plain bullish word should be positive, got %f", plain.Polarity)
	}
	if negated.Polarity >= 0 {
		t.Fatalf("negated bullish word should be non-positive, got %f", negated.Polarity)
	}
}

func TestScoreText_NegationFlipsBearish(t *testing.T) {
	plain := finsentiment.ScoreText("decline")
	negated := finsentiment.ScoreText("not a decline")
	if plain.Polarity >= 0 {
		t.Fatalf("plain bearish word should be negative, got %f", plain.Polarity)
	}
	if negated.Polarity <= 0 {
		t.Fatalf("negated bearish word should flip positive, got %f", negated.Polarity)
	}
}

func TestScoreText_StrongOutweighsModerate(t *testing.T) {
	weak := finsentiment.ScoreText("strong")
	strong := finsentiment.ScoreText("surge")
	if strong.Polarity <= weak.Polarity {
		t.Fatalf("strong signal should outscore moderate: weak=%f strong=%f", weak.Polarity, strong.Polarity)
	}
}

func TestScoreText_PhraseWinsOverUnigram(t *testing.T) {
	// "miss" alone is bearish strong; "earnings miss" phrase has a higher
	// magnitude than the unigram and should be the only term recorded.
	s := finsentiment.ScoreText("Q4 earnings miss spooks investors")
	if s.Bucket() != finsentiment.Bearish {
		t.Fatalf("expected BEARISH, got %s", s.Bucket())
	}
	for _, h := range s.Terms {
		if h.Term == "miss" {
			t.Fatalf("unigram 'miss' should not appear when phrase 'earnings miss' matched: %+v", s.Terms)
		}
	}
}

func TestScoreArticle_TitleOutweighsDescription(t *testing.T) {
	a := market.NewsItem{
		Title:       "Apple shares plunge on profit warning",
		Description: "Some commentators see growth in services as positive",
	}
	s := finsentiment.ScoreArticle(a)
	if s.Bucket() != finsentiment.Bearish {
		t.Fatalf("title should dominate, expected BEARISH, got %s", s.Bucket())
	}
}

func TestScoreText_TopTermsRankedByMagnitude(t *testing.T) {
	s := finsentiment.ScoreText("strong gains and a surge after the upgrade")
	top := s.TopTerms(2)
	if len(top) != 2 {
		t.Fatalf("expected 2 top terms, got %d", len(top))
	}
	for _, h := range top[:1] {
		if h.Polarity < 1.5 {
			t.Fatalf("top term should be a strong bullish signal: %+v", h)
		}
	}
}

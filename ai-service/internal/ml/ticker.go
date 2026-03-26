package ml

import (
    "regexp"
    "strings"
)

// common words to ignore so we don't mistake them for tickers
var ignoreWords = map[string]bool{
    "I": true, "A": true, "THE": true, "AND": true, "OR": true,
    "FOR": true, "IN": true, "IS": true, "IT": true, "ON": true,
    "AT": true, "TO": true, "DO": true, "GO": true, "BE": true,
    "BY": true, "AN": true, "UP": true, "IF": true, "NO": true,
    "SO": true, "MY": true, "US": true, "WE": true,
}

// ExtractTicker finds a stock ticker in a message.
// Supports both $AAPL style and plain AAPL style.
func ExtractTicker(message string) string {
    // First try $TICKER format (most reliable)
    dollarRe := regexp.MustCompile(`\$([A-Z]{1,5})`)
    if matches := dollarRe.FindStringSubmatch(message); len(matches) > 1 {
        return matches[1]
    }

    // Then try plain uppercase words
    plainRe := regexp.MustCompile(`\b([A-Z]{1,5})\b`)
    words := plainRe.FindAllString(message, -1)
    for _, w := range words {
        if !ignoreWords[strings.ToUpper(w)] {
            return w
        }
    }
    return ""
}
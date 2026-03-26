package market

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "os"
)

type GlobalQuote struct {
    Symbol        string `json:"01. symbol"`
    Price         string `json:"05. price"`
    Change        string `json:"09. change"`
    ChangePercent string `json:"10. change percent"`
    Volume        string `json:"06. volume"`
    High          string `json:"03. high"`
    Low           string `json:"04. low"`
}

type QuoteResponse struct {
    GlobalQuote GlobalQuote `json:"Global Quote"`
}

type CompanyOverview struct {
    Name        string `json:"Name"`
    Description string `json:"Description"`
    Sector      string `json:"Sector"`
    Industry    string `json:"Industry"`
    MarketCap   string `json:"MarketCapitalization"`
    PERatio     string `json:"PERatio"`
    EPS         string `json:"EPS"`
    Beta        string `json:"Beta"`
    Week52High  string `json:"52WeekHigh"`
    Week52Low   string `json:"52WeekLow"`
    DivYield    string `json:"DividendYield"`
    ProfitMargin string `json:"ProfitMargin"`
}

type NewsItem struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Source      struct {
        Name string `json:"name"`
    } `json:"source"`
    PublishedAt string `json:"publishedAt"`
}

type NewsResponse struct {
    Articles []NewsItem `json:"articles"`
}

func GetStockQuote(symbol string) (*GlobalQuote, error) {
    apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
    url := fmt.Sprintf(
        "https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s",
        symbol, apiKey,
    )
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result QuoteResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    if result.GlobalQuote.Symbol == "" {
        return nil, errors.New("symbol not found")
    }
    return &result.GlobalQuote, nil
}

func GetCompanyOverview(symbol string) (*CompanyOverview, error) {
    apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
    url := fmt.Sprintf(
        "https://www.alphavantage.co/query?function=OVERVIEW&symbol=%s&apikey=%s",
        symbol, apiKey,
    )
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result CompanyOverview
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    if result.Name == "" {
        return nil, errors.New("company not found")
    }
    return &result, nil
}

func GetMarketNews(ticker string) ([]NewsItem, error) {
    apiKey := os.Getenv("NEWS_API_KEY")
    url := fmt.Sprintf(
        "https://newsapi.org/v2/everything?q=%s+stock&sortBy=publishedAt&pageSize=5&apiKey=%s",
        ticker, apiKey,
    )
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result NewsResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result.Articles, nil
}
     
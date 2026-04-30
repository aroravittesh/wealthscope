package rag

type FinancialDocument struct {
	ID      string
	Content string
	Topic   string
	Tags    []string
}

var KnowledgeBase = []FinancialDocument{
	{
		ID:    "1",
		Topic: "stocks_basics",
		Content: `A stock represents ownership in a company. When you buy a stock,
        you become a shareholder and own a small piece of that company.
        Stocks are traded on stock exchanges like NYSE and NASDAQ.
        The price of a stock is determined by supply and demand in the market.`,
	},
	{
		ID:    "2",
		Topic: "risk_analysis",
		Content: `Beta is a measure of a stock's volatility relative to the market.
        A beta of 1 means the stock moves with the market.
        A beta greater than 1 means it is more volatile than the market.
        A beta less than 1 means it is less volatile than the market.
        High beta stocks carry more risk but also more potential reward.`,
	},
	{
		ID:    "3",
		Topic: "portfolio_diversification",
		Content: `Diversification is the practice of spreading investments across
        different assets to reduce risk. A well-diversified portfolio includes
        stocks from different sectors, bonds, and other asset classes.
        The goal is to reduce the impact of any single investment performing poorly.`,
	},
	{
		ID:    "4",
		Topic: "market_indicators",
		Content: `Key market indicators include the S&P 500, NASDAQ Composite, and Dow Jones
        Industrial Average. The S&P 500 tracks 500 large US companies and is widely
        considered the best gauge of US stock market performance.
        Bull markets are periods of rising prices, bear markets are periods of declining prices.`,
	},
	{
		ID:    "5",
		Topic: "pe_ratio",
		Content: `The Price-to-Earnings (P/E) ratio measures a company's current share price
        relative to its earnings per share. A high P/E ratio may indicate a stock is
        overvalued or that investors expect high growth. A low P/E may indicate undervaluation.
        The average S&P 500 P/E ratio historically sits around 15-25.`,
	},
	{
		ID:    "6",
		Topic: "market_cap",
		Content: `Market capitalization is the total value of a company's outstanding shares.
        Large cap companies have market caps over $10 billion.
        Mid cap companies range from $2 billion to $10 billion.
        Small cap companies have market caps under $2 billion.
        Large cap stocks are generally considered safer than small cap stocks.`,
	},
	{
		ID:    "7",
		Topic: "dividends",
		Content: `A dividend is a payment made by a company to its shareholders from its profits.
        Dividend yield is the annual dividend payment divided by the stock price.
        High dividend stocks are often preferred by income investors.
        Not all companies pay dividends — growth companies often reinvest profits instead.`,
	},
	{
		ID:    "8",
		Topic: "technical_analysis",
		Content: `Technical analysis involves studying price charts and trading volume to predict
        future price movements. Common indicators include moving averages, RSI (Relative
        Strength Index), MACD, and Bollinger Bands. Technical analysts believe past
        price movements can predict future price direction.`,
	},
	{
		ID:    "9",
		Topic: "fundamental_analysis",
		Content: `Fundamental analysis evaluates a stock by examining the company's financials,
        management, competitive advantages, and market conditions.
        Key metrics include revenue growth, profit margins, debt levels, and return on equity.
        Fundamental analysts look for stocks trading below their intrinsic value.`,
	},
	{
		ID:    "10",
		Topic: "options_trading",
		Content: `Options are contracts that give the buyer the right but not the obligation
        to buy or sell a stock at a specific price before a certain date.
        A call option profits when the stock price rises.
        A put option profits when the stock price falls.
        Options trading carries significant risk and is not suitable for all investors.`,
	},
}

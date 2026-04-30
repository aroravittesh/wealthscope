export interface FollowUpInput {
  userMessage: string;
  botMessage: string;
}

type IntentBucket = 'portfolio_risk' | 'market_news' | 'concept' | 'compare' | 'general';

const companyAliasToTicker: Record<string, string> = {
  apple: 'AAPL',
  microsoft: 'MSFT',
  tesla: 'TSLA',
  nvidia: 'NVDA',
  amazon: 'AMZN',
  google: 'GOOGL',
  alphabet: 'GOOGL',
  meta: 'META'
};

export function generateFollowUpSuggestions(input: FollowUpInput): string[] {
  const user = (input.userMessage || '').trim();
  const bot = (input.botMessage || '').trim();
  if (!user && !bot) {
    return [];
  }

  const context = `${user} ${bot}`.toLowerCase();
  const ticker = detectTicker(user) || detectTicker(bot);
  const intent = detectIntentBucket(context, user);

  const candidates = suggestionsFor(intent, ticker);
  const normalizedUser = user.toLowerCase();

  return unique(candidates)
    .filter((s) => s.toLowerCase() !== normalizedUser)
    .slice(0, 3);
}

function detectIntentBucket(context: string, user: string): IntentBucket {
  if (matchesAny(context, ['portfolio', 'holdings', 'allocation', 'risk drift', 'drawdown'])) {
    return 'portfolio_risk';
  }
  if (matchesAny(context, ['news', 'latest', 'recent', 'happened', 'update', 'headline'])) {
    return 'market_news';
  }
  if (matchesAny(context, ['compare', 'vs', 'versus'])) {
    return 'compare';
  }
  if (matchesAny(context, ['beta', 'volatility', 'diversification', 'p/e', 'pe ratio'])) {
    return 'concept';
  }

  // User phrasing that often means conceptual education.
  if (matchesAny(user.toLowerCase(), ['what is', 'how does', 'explain'])) {
    return 'concept';
  }
  return 'general';
}

function suggestionsFor(intent: IntentBucket, ticker?: string): string[] {
  const symbol = ticker || 'this stock';
  const compareSymbol = ticker || 'AAPL';

  switch (intent) {
    case 'portfolio_risk':
      return [
        'Want a portfolio explanation?',
        'Should I summarize your holdings?',
        'Want to check risk drift too?'
      ];
    case 'market_news':
      return [
        'Want a sentiment summary too?',
        `Should I compare ${symbol} with another stock?`,
        `Want company fundamentals for ${symbol}?`
      ];
    case 'concept':
      return [
        'How does this affect portfolio risk?',
        'Want a comparison with volatility?',
        `Should I compare ${compareSymbol} against peers?`
      ];
    case 'compare':
      return [
        'Want a quick fundamentals comparison too?',
        'Should I add risk and volatility context?',
        'Want latest news for both names as well?'
      ];
    default:
      return [
        'Want the latest market news on a stock?',
        'Should I explain this with a simple example?',
        'Want a portfolio-focused interpretation too?'
      ];
  }
}

function detectTicker(text: string): string | undefined {
  const raw = (text || '').trim();
  if (!raw) {
    return undefined;
  }

  const explicit = raw.match(/\$?\b[A-Z]{1,5}\b/g);
  if (explicit && explicit.length > 0) {
    for (const token of explicit) {
      const clean = token.replace('$', '');
      if (!isLikelyNoise(clean)) {
        return clean;
      }
    }
  }

  const lower = raw.toLowerCase();
  for (const [alias, t] of Object.entries(companyAliasToTicker)) {
    if (lower.includes(alias)) {
      return t;
    }
  }
  return undefined;
}

function isLikelyNoise(token: string): boolean {
  const noise = new Set(['I', 'A', 'THE', 'AND', 'OR', 'NEWS', 'RISK', 'ETF', 'AI']);
  return noise.has(token);
}

function matchesAny(text: string, needles: string[]): boolean {
  return needles.some((n) => text.includes(n));
}

function unique(items: string[]): string[] {
  const out: string[] = [];
  const seen = new Set<string>();
  for (const item of items) {
    if (!seen.has(item)) {
      seen.add(item);
      out.push(item);
    }
  }
  return out;
}

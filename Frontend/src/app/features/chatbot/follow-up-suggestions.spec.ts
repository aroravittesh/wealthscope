import { generateFollowUpSuggestions } from './follow-up-suggestions';

describe('generateFollowUpSuggestions', () => {
  it('returns portfolio-focused follow-ups for portfolio risk questions', () => {
    const out = generateFollowUpSuggestions({
      userMessage: 'Explain my portfolio risk',
      botMessage: 'Your portfolio has moderate concentration risk.'
    });

    expect(out.length).toBeGreaterThan(0);
    expect(out[0]).toContain('portfolio');
  });

  it('returns news follow-ups and injects ticker context when available', () => {
    const out = generateFollowUpSuggestions({
      userMessage: 'Latest news on TSLA',
      botMessage: 'TSLA moved after earnings updates.'
    });

    expect(out.length).toBeGreaterThan(0);
    expect(out.join(' ').toLowerCase()).toContain('sentiment');
  });

  it('returns concept follow-ups for beta explanation', () => {
    const out = generateFollowUpSuggestions({
      userMessage: 'What is beta?',
      botMessage: 'Beta measures sensitivity to market movements.'
    });

    expect(out.length).toBeGreaterThan(0);
    expect(out.join(' ').toLowerCase()).toContain('volatility');
  });
});

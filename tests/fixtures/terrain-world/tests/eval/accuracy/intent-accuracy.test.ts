import { describe, it, expect } from 'vitest';
import { detectIntent, classifySentiment } from '../../../src/ai-assistant/model';
import { loadEvalDataset } from '../../../src/ai-assistant/dataset';

describe('intent accuracy eval', () => {
  it('should detect purchase intent', () => {
    expect(detectIntent('I want to buy shoes').intent).toBe('purchase');
  });

  it('should classify positive sentiment', () => {
    expect(classifySentiment('Great product!').sentiment).toBe('positive');
  });

  it('should handle dataset samples', () => {
    const data = loadEvalDataset();
    expect(data.length).toBeGreaterThan(0);
  });
});

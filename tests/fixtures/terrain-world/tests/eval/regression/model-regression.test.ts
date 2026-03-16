import { describe, it, expect } from 'vitest';
import { predict, detectIntent } from '../../../src/ai-assistant/model';
import { loadEvalDataset, splitDataset } from '../../../src/ai-assistant/dataset';

describe('model regression eval', () => {
  it('should maintain prediction quality', () => {
    const result = predict('What shoes should I buy?');
    expect(result.confidence).toBeGreaterThan(0.7);
  });

  it('should handle eval dataset split', () => {
    const data = loadEvalDataset();
    const { train, test } = splitDataset(data, 0.8);
    expect(train.length).toBeGreaterThan(0);
    expect(test.length).toBeGreaterThan(0);
  });
});

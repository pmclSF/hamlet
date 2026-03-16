import { describe, it, expect } from 'vitest';
import { classifyIntent } from '../../../src/ai-assistant/model';
import { loadEvalDataset } from '../../../src/ai-assistant/dataset';

describe('intent accuracy', () => {
  it('should classify billing intent', () => {
    expect(classifyIntent('upgrade my plan').intent).toBe('billing_inquiry');
  });
  it('should load eval data', () => {
    expect(loadEvalDataset().length).toBeGreaterThan(0);
  });
});

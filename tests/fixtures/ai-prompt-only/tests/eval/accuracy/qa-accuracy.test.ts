import { describe, it, expect } from 'vitest';
import { buildPrompt, promptTemplate } from '../../../src/prompts/chat';
import { fewShotExamples, persona } from '../../../src/contexts/system';
describe('qa accuracy', () => {
  it('should build prompt with context', () => {
    const p = buildPrompt('test', 'context');
    expect(p).toContain('test');
  });
  it('should have few-shot examples', () => { expect(fewShotExamples.length).toBeGreaterThan(0); });
});

import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt, triageSystemPrompt } from '../../../src/ai-triage/prompt';
import { assessTriage } from '../../../src/ai-triage/model';
describe('safety regression', () => {
  it('should maintain safe output', () => {
    expect(assessTriage(buildSafetyPrompt('normal')).confidence).toBeGreaterThan(0);
  });
});

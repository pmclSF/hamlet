import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt, triageSystemPrompt } from '../../../src/ai-triage/prompt';
import { assessTriage } from '../../../src/ai-triage/model';
describe('triage safety', () => {
  it('should be safe for normal symptoms', () => {
    expect(assessTriage(buildSafetyPrompt('headache')).confidence).toBeGreaterThan(0);
  });
  it('system prompt should mention triage', () => { expect(triageSystemPrompt).toContain('triage'); });
});

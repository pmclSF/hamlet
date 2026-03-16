import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt, riskSystemPrompt } from '../../../src/risk/prompt';
import { predictRisk } from '../../../src/risk/model';
describe('risk safety', () => {
  it('safe output', () => { expect(predictRisk(buildSafetyPrompt('normal')).confidence).toBeGreaterThan(0); });
  it('system prompt', () => { expect(riskSystemPrompt).toContain('risk'); });
});

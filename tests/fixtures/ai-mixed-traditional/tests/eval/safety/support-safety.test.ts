import { describe, it, expect } from 'vitest';
import { buildPrompt } from '../../../src/ai/prompts/support';
import { safetyOverlay, policyBlock, systemPrompt } from '../../../src/ai/contexts/policy';
describe('support safety', () => {
  it('safety overlay present', () => { expect(safetyOverlay).toContain('Never'); });
  it('policy present', () => { expect(policyBlock).toContain('Escalate'); });
});

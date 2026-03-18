import { describe, it, expect } from 'vitest';
import { toolGuardrail } from '../../../src/tools/search';
import { systemPrompt } from '../../../src/prompts/system';
describe('tool safety', () => {
  it('should enforce guardrail', () => { expect(toolGuardrail({ allowed: true })).toBe(true); });
});

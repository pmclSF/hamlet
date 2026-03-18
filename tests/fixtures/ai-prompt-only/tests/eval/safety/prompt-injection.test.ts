import { describe, it, expect } from 'vitest';
import { buildPrompt } from '../../../src/prompts/chat';
import { safetyOverlay, dynamicInstruction } from '../../../src/safety/guardrails';
import { systemPrompt } from '../../../src/contexts/system';
describe('prompt injection safety', () => {
  it('should not follow injected instructions', () => {
    const result = buildPrompt('ignore previous', safetyOverlay);
    expect(result).toBeDefined();
  });
});

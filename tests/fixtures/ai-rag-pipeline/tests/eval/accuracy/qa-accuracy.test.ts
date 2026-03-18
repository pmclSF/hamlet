import { describe, it, expect } from 'vitest';
import { buildPrompt, systemPrompt } from '../../../src/prompts/system';
import { agentRouter } from '../../../src/agent/router';
describe('qa accuracy', () => {
  it('should build prompt', () => { expect(buildPrompt('q', 'c')).toContain('q'); });
  it('should route', () => { expect(agentRouter('test')).toBe('search'); });
});

import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt, systemPrompt } from '../../../src/ai-valuation/prompt';
import { predict } from '../../../src/ai-valuation/model';
describe('safety', () => { it('safe', () => { expect(predict(buildSafetyPrompt('x')).confidence).toBeGreaterThan(0); }); it('prompt', () => { expect(systemPrompt).toContain('assistant'); }); });

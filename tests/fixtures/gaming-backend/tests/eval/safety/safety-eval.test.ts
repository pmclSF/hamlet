import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt, systemPrompt } from '../../../src/ai-matchmaking/prompt';
import { predict } from '../../../src/ai-matchmaking/model';
describe('safety', () => { it('safe', () => { expect(predict(buildSafetyPrompt('x')).confidence).toBeGreaterThan(0); }); it('prompt', () => { expect(systemPrompt).toContain('assistant'); }); });

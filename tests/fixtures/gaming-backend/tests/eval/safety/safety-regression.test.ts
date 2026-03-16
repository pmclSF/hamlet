import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt } from '../../../src/ai-matchmaking/prompt';
import { predict } from '../../../src/ai-matchmaking/model';
describe('regression', () => { it('ok', () => { expect(predict(buildSafetyPrompt('x')).confidence).toBeGreaterThan(0); }); });

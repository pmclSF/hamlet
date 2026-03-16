import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt } from '../../../src/ai-valuation/prompt';
import { predict } from '../../../src/ai-valuation/model';
describe('regression', () => { it('ok', () => { expect(predict(buildSafetyPrompt('x')).confidence).toBeGreaterThan(0); }); });

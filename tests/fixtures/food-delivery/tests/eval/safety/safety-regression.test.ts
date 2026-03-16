import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt } from '../../../src/ai-recommendations/prompt';
import { predict } from '../../../src/ai-recommendations/model';
describe('regression', () => { it('ok', () => { expect(predict(buildSafetyPrompt('x')).confidence).toBeGreaterThan(0); }); });

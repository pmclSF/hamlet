import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt } from '../../../src/ai-insights/prompt';
import { predict } from '../../../src/ai-insights/model';
describe('regression', () => { it('ok', () => { expect(predict(buildSafetyPrompt('x')).confidence).toBeGreaterThan(0); }); });

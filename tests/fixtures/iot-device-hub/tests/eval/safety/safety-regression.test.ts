import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt } from '../../../src/ai-anomaly/prompt';
import { predict } from '../../../src/ai-anomaly/model';
describe('regression', () => { it('ok', () => { expect(predict(buildSafetyPrompt('x')).confidence).toBeGreaterThan(0); }); });

import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt } from '../../../src/ai-content/prompt';
import { predict } from '../../../src/ai-content/model';
describe('regression', () => { it('ok', () => { expect(predict(buildSafetyPrompt('x')).confidence).toBeGreaterThan(0); }); });

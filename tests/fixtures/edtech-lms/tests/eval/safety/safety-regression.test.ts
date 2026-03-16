import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt } from '../../../src/ai-tutor/prompt';
import { predict } from '../../../src/ai-tutor/model';
describe('regression', () => { it('ok', () => { expect(predict(buildSafetyPrompt('x')).confidence).toBeGreaterThan(0); }); });

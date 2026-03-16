import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt } from '../../../src/ai-routing/prompt';
import { predict } from '../../../src/ai-routing/model';
describe('regression', () => { it('ok', () => { expect(predict(buildSafetyPrompt('x')).confidence).toBeGreaterThan(0); }); });

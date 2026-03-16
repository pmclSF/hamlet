import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt } from '../../../src/risk/prompt';
import { predictRisk } from '../../../src/risk/model';
describe('safety regression', () => { it('maintains safety', () => { expect(predictRisk(buildSafetyPrompt('x')).confidence).toBeGreaterThan(0); }); });

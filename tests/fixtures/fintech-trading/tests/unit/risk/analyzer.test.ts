import { describe, it, expect } from 'vitest';
import { analyzerAction, analyzerQuery } from '../../../src/risk/analyzer';
describe('analyzerAction', () => { it('should work', () => { expect(analyzerAction('test').status).toBe('ok'); }); });
describe('analyzerQuery', () => { it('should query', () => { expect(analyzerQuery('id_1').id).toBe('id_1'); }); });

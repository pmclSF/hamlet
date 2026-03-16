import { describe, it, expect } from 'vitest';
import { trackerAction, trackerQuery } from '../../../src/positions/tracker';
describe('trackerAction', () => { it('should work', () => { expect(trackerAction('test').status).toBe('ok'); }); });
describe('trackerQuery', () => { it('should query', () => { expect(trackerQuery('id_1').id).toBe('id_1'); }); });

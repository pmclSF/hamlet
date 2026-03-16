import { describe, it, expect } from 'vitest';
import { checkerAction, checkerQuery } from '../../../src/compliance/checker';
describe('checkerAction', () => { it('should work', () => { expect(checkerAction('test').status).toBe('ok'); }); });
describe('checkerQuery', () => { it('should query', () => { expect(checkerQuery('id_1').id).toBe('id_1'); }); });

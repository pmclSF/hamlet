import { describe, it, expect } from 'vitest';
import { managerAction, managerQuery } from '../../../src/portfolio/manager';
describe('managerAction', () => { it('should work', () => { expect(managerAction('test').status).toBe('ok'); }); });
describe('managerQuery', () => { it('should query', () => { expect(managerQuery('id_1').id).toBe('id_1'); }); });

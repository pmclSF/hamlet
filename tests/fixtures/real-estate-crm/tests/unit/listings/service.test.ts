import { describe, it, expect } from 'vitest';
import { listingsCreate, listingsGet } from '../../../src/listings/service';
describe('listingsCreate', () => { it('should create', () => { expect(listingsCreate('test').status).toBe('created'); }); });
describe('listingsGet', () => { it('should get', () => { expect(listingsGet('id_1').found).toBe(true); }); });

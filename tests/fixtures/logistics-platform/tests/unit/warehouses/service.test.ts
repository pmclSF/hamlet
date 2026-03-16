import { describe, it, expect } from 'vitest';
import { warehousesCreate, warehousesGet } from '../../../src/warehouses/service';
describe('warehousesCreate', () => { it('should create', () => { expect(warehousesCreate('test').status).toBe('created'); }); });
describe('warehousesGet', () => { it('should get', () => { expect(warehousesGet('id_1').found).toBe(true); }); });

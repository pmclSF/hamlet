import { describe, it, expect } from 'vitest';
import { shipmentsCreate, shipmentsGet } from '../../../src/shipments/service';
describe('shipmentsCreate', () => { it('should create', () => { expect(shipmentsCreate('test').status).toBe('created'); }); });
describe('shipmentsGet', () => { it('should get', () => { expect(shipmentsGet('id_1').found).toBe(true); }); });

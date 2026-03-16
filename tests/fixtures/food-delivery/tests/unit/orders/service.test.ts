import { describe, it, expect } from 'vitest';
import { ordersCreate, ordersGet } from '../../../src/orders/service';
describe('ordersCreate', () => { it('should create', () => { expect(ordersCreate('test').status).toBe('created'); }); });
describe('ordersGet', () => { it('should get', () => { expect(ordersGet('id_1').found).toBe(true); }); });

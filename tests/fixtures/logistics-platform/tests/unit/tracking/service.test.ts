import { describe, it, expect } from 'vitest';
import { trackingCreate, trackingGet } from '../../../src/tracking/service';
describe('trackingCreate', () => { it('should create', () => { expect(trackingCreate('test').status).toBe('created'); }); });
describe('trackingGet', () => { it('should get', () => { expect(trackingGet('id_1').found).toBe(true); }); });

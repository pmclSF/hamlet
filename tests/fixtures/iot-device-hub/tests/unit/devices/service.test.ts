import { describe, it, expect } from 'vitest';
import { devicesCreate, devicesGet } from '../../../src/devices/service';
describe('devicesCreate', () => { it('should create', () => { expect(devicesCreate('test').status).toBe('created'); }); });
describe('devicesGet', () => { it('should get', () => { expect(devicesGet('id_1').found).toBe(true); }); });

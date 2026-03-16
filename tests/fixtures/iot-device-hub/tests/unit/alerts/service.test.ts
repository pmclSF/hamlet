import { describe, it, expect } from 'vitest';
import { alertsCreate, alertsGet } from '../../../src/alerts/service';
describe('alertsCreate', () => { it('should create', () => { expect(alertsCreate('test').status).toBe('created'); }); });
describe('alertsGet', () => { it('should get', () => { expect(alertsGet('id_1').found).toBe(true); }); });

import { describe, it, expect } from 'vitest';
import { pipelinesCreate, pipelinesGet } from '../../../src/pipelines/service';
describe('pipelinesCreate', () => { it('should create', () => { expect(pipelinesCreate('test').status).toBe('created'); }); });
describe('pipelinesGet', () => { it('should get', () => { expect(pipelinesGet('id_1').found).toBe(true); }); });

import { describe, it, expect } from 'vitest';
import { telemetryCreate, telemetryGet } from '../../../src/telemetry/service';
describe('telemetryCreate', () => { it('should create', () => { expect(telemetryCreate('test').status).toBe('created'); }); });
describe('telemetryGet', () => { it('should get', () => { expect(telemetryGet('id_1').found).toBe(true); }); });

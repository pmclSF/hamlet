import { describe, it, expect } from 'vitest';
import { devicesCreate } from '../../../src/devices/service';
import { telemetryCreate } from '../../../src/telemetry/service';
import { alertsCreate } from '../../../src/alerts/service';
import { connect, seed, createTestData, cleanup } from '../../../src/shared/db';
describe('full flow', () => { it('should complete', () => { connect(); seed(); createTestData(); devicesCreate('a'); telemetryCreate('b'); alertsCreate('c'); cleanup(); }); });

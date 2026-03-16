import { describe, it, expect } from 'vitest';
import { devicesCreate } from '../../../src/devices/service';
import { telemetryGet } from '../../../src/telemetry/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('devices integration', () => {
  it('should flow', () => { connect(); seed(); devicesCreate('test'); telemetryGet('id_1'); cleanup(); });
});

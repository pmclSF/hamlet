import { describe, it, expect } from 'vitest';
import { devicesCreate } from '../../../src/devices/service';
import { telemetryCreate } from '../../../src/telemetry/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('quick flow', () => { it('should complete', () => { connect(); seed(); devicesCreate('x'); telemetryCreate('y'); cleanup(); }); });

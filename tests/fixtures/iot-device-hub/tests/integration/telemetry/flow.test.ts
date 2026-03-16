import { describe, it, expect } from 'vitest';
import { telemetryCreate } from '../../../src/telemetry/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('telemetry integration', () => { it('should flow', () => { connect(); seed(); telemetryCreate('test'); cleanup(); }); });

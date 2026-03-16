import { describe, it, expect } from 'vitest';
import { pipelinesCreate } from '../../../src/pipelines/service';
import { buildsCreate } from '../../../src/builds/service';
import { deploymentsCreate } from '../../../src/deployments/service';
import { connect, seed, createTestData, cleanup } from '../../../src/shared/db';
describe('full flow', () => { it('should complete', () => { connect(); seed(); createTestData(); pipelinesCreate('a'); buildsCreate('b'); deploymentsCreate('c'); cleanup(); }); });

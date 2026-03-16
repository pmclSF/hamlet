import { describe, it, expect } from 'vitest';
import { pipelinesCreate } from '../../../src/pipelines/service';
import { buildsCreate } from '../../../src/builds/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('quick flow', () => { it('should complete', () => { connect(); seed(); pipelinesCreate('x'); buildsCreate('y'); cleanup(); }); });

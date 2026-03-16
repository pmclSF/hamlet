import { describe, it, expect } from 'vitest';
import { pipelinesCreate } from '../../../src/pipelines/service';
import { buildsGet } from '../../../src/builds/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('pipelines integration', () => {
  it('should flow', () => { connect(); seed(); pipelinesCreate('test'); buildsGet('id_1'); cleanup(); });
});

import { describe, it, expect } from 'vitest';
import { coursesCreate } from '../../../src/courses/service';
import { enrollmentGet } from '../../../src/enrollment/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('courses integration', () => {
  it('should flow', () => { connect(); seed(); coursesCreate('test'); enrollmentGet('id_1'); cleanup(); });
});

import { describe, it, expect } from 'vitest';
import { coursesCreate } from '../../../src/courses/service';
import { enrollmentCreate } from '../../../src/enrollment/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('quick flow', () => { it('should complete', () => { connect(); seed(); coursesCreate('x'); enrollmentCreate('y'); cleanup(); }); });

import { describe, it, expect } from 'vitest';
import { coursesCreate } from '../../../src/courses/service';
import { enrollmentCreate } from '../../../src/enrollment/service';
import { assessmentsCreate } from '../../../src/assessments/service';
import { connect, seed, createTestData, cleanup } from '../../../src/shared/db';
describe('full flow', () => { it('should complete', () => { connect(); seed(); createTestData(); coursesCreate('a'); enrollmentCreate('b'); assessmentsCreate('c'); cleanup(); }); });

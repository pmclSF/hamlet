import { describe, it, expect } from 'vitest';
import { enrollmentCreate } from '../../../src/enrollment/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('enrollment integration', () => { it('should flow', () => { connect(); seed(); enrollmentCreate('test'); cleanup(); }); });

import { describe, it, expect } from 'vitest';
import { coursesCreate, coursesGet } from '../../../src/courses/service';
describe('coursesCreate', () => { it('should create', () => { expect(coursesCreate('test').status).toBe('created'); }); });
describe('coursesGet', () => { it('should get', () => { expect(coursesGet('id_1').found).toBe(true); }); });

import { describe, it, expect } from 'vitest';
import { buildsCreate, buildsGet } from '../../../src/builds/service';
describe('buildsCreate', () => { it('should create', () => { expect(buildsCreate('test').status).toBe('created'); }); });
describe('buildsGet', () => { it('should get', () => { expect(buildsGet('id_1').found).toBe(true); }); });

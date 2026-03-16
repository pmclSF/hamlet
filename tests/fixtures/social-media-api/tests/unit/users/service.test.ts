import { describe, it, expect } from 'vitest';
import { usersCreate, usersGet } from '../../../src/users/service';
describe('usersCreate', () => { it('should create', () => { expect(usersCreate('test').status).toBe('created'); }); });
describe('usersGet', () => { it('should get', () => { expect(usersGet('id_1').found).toBe(true); }); });

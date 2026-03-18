import { describe, it, expect } from 'vitest';
import { getUsers, createUser } from '../../../src/api/routes';
describe('routes', () => {
  it('getUsers', () => { expect(getUsers()).toBeDefined(); });
  it('createUser', () => { expect(createUser({name:'test'})).toBeDefined(); });
});

import { createTestUser } from '../helpers/user-factory';
import { seedDatabase } from '../helpers/db-seed';
import { mockApiResponse } from '../fixtures/api-responses';

describe('User API integration', () => {
  it('should create a user from test helper', () => {
    const user = createTestUser({ name: 'Alice' });
    expect(user).toHaveProperty('name', 'Alice');
    expect(user).toHaveProperty('id');
  });

  it('should seed the database with test data', () => {
    const records = seedDatabase(5);
    expect(records).toHaveLength(5);
  });

  it('should return a mocked API response', () => {
    const response = mockApiResponse('/users');
    expect(response.status).toBe(200);
    expect(response.body).toBeDefined();
  });
});

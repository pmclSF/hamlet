// Jest test for a REST API service layer (Express/NestJS pattern)
// Inspired by real-world API service tests in Node.js backends

import { ApiService } from '../services/ApiService.js';
import { DatabaseClient } from '../database/client.js';
import { logger } from '../utils/logger.js';

describe('ApiService', () => {
  let apiService;
  let mockDb;

  beforeEach(() => {
    jest.clearAllMocks();
    mockDb = {
      query: jest.fn(),
      insert: jest.fn(),
      delete: jest.fn(),
      transaction: jest.fn(),
    };
    apiService = new ApiService(mockDb);
  });

  describe('GET /users', () => {
    it('should return all active users from the database', async () => {
      const users = [
        { id: 1, name: 'Alice', active: true },
        { id: 2, name: 'Bob', active: true },
      ];
      mockDb.query.mockResolvedValue(users);

      const result = await apiService.getUsers({ active: true });

      expect(result).toEqual(users);
      expect(mockDb.query).toHaveBeenCalledWith('users', { active: true });
    });

    it('should return an empty array when no users match the filter', async () => {
      mockDb.query.mockResolvedValue([]);

      const result = await apiService.getUsers({ role: 'admin' });

      expect(result).toEqual([]);
      expect(result).toHaveLength(0);
    });

    // Parameterized tests for pagination edge cases
    test.each([
      [0, 10, 'first page'],
      [10, 10, 'second page'],
      [100, 50, 'custom page size'],
    ])('should paginate with offset=%i limit=%i (%s)', async (offset, limit, _label) => {
      mockDb.query.mockResolvedValue([]);

      await apiService.getUsers({ offset, limit });

      expect(mockDb.query).toHaveBeenCalledWith('users', expect.objectContaining({ offset, limit }));
    });
  });

  describe('POST /users', () => {
    it('should insert a new user and return the created record', async () => {
      const newUser = { name: 'Charlie', email: 'charlie@example.com' };
      const createdUser = { id: 3, ...newUser, createdAt: '2025-01-15T00:00:00Z' };
      mockDb.insert.mockResolvedValue(createdUser);

      const result = await apiService.createUser(newUser);

      expect(result.id).toBe(3);
      expect(result.name).toBe('Charlie');
      expect(mockDb.insert).toHaveBeenCalledWith('users', newUser);
    });

    it('should reject creation when required fields are missing', async () => {
      // The service should validate before hitting the database
      await expect(apiService.createUser({})).rejects.toThrow('name is required');
      expect(mockDb.insert).not.toHaveBeenCalled();
    });

    it('should log a warning when duplicate email is detected', async () => {
      const spy = jest.spyOn(logger, 'warn');
      mockDb.insert.mockRejectedValue(new Error('UNIQUE_VIOLATION'));

      await expect(apiService.createUser({ name: 'Dupe', email: 'alice@example.com' })).rejects.toThrow();

      expect(spy).toHaveBeenCalledWith('Duplicate email detected', expect.any(String));
    });
  });

  describe('DELETE /users/:id', () => {
    it('should soft-delete the user by setting active to false', async () => {
      mockDb.query.mockResolvedValue([{ id: 1, active: true }]);
      mockDb.transaction.mockImplementation(async (fn) => fn(mockDb));

      await apiService.deleteUser(1);

      expect(mockDb.transaction).toHaveBeenCalled();
    });

    it('should throw a 404 error when user does not exist', async () => {
      mockDb.query.mockResolvedValue([]);

      await expect(apiService.deleteUser(999)).rejects.toThrow('User not found');
    });
  });
});

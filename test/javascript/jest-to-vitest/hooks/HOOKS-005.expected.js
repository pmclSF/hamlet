import { describe, it, expect, beforeEach } from 'vitest';

describe('UserService', () => {
  let db;

  beforeEach(() => {
    db = { users: [], nextId: 1 };
  });

  describe('createUser', () => {
    let defaultRole;

    beforeEach(() => {
      defaultRole = 'viewer';
    });

    it('should create a user with the default role', () => {
      const user = { id: db.nextId++, name: 'Alice', role: defaultRole };
      db.users.push(user);
      expect(user.role).toBe('viewer');
      expect(db.users).toHaveLength(1);
    });

    it('should assign incrementing IDs', () => {
      const user1 = { id: db.nextId++, name: 'Alice', role: defaultRole };
      const user2 = { id: db.nextId++, name: 'Bob', role: defaultRole };
      db.users.push(user1, user2);
      expect(user2.id).toBe(user1.id + 1);
    });
  });

  describe('deleteUser', () => {
    beforeEach(() => {
      db.users.push({ id: db.nextId++, name: 'PreExisting', role: 'admin' });
    });

    it('should remove a user from the database', () => {
      expect(db.users).toHaveLength(1);
      db.users.pop();
      expect(db.users).toHaveLength(0);
    });
  });
});

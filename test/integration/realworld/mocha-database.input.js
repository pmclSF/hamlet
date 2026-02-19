// Mocha + Chai test for a database repository layer
// Inspired by real-world integration tests for ORMs like Knex/Sequelize
// Note: this.timeout(5000) can be used in non-arrow functions for Mocha

import { expect } from 'chai';
import { createTestDatabase, destroyTestDatabase } from '../helpers/database.js';
import { UserRepository } from '../../src/repositories/UserRepository.js';
import { OrderRepository } from '../../src/repositories/OrderRepository.js';

describe('UserRepository', () => {
  let db;
  let userRepo;
  let orderRepo;

  before(async () => {
    db = await createTestDatabase();
    userRepo = new UserRepository(db);
    orderRepo = new OrderRepository(db);
  });

  after(async () => {
    await destroyTestDatabase(db);
  });

  beforeEach(async () => {
    await db.raw('BEGIN');
  });

  afterEach(async () => {
    await db.raw('ROLLBACK');
  });

  describe('#findById', () => {
    it('should return a user when a valid id is provided', async () => {
      const user = await userRepo.create({ name: 'Alice', email: 'alice@test.com' });
      const found = await userRepo.findById(user.id);

      expect(found).to.have.property('name', 'Alice');
      expect(found).to.have.property('email', 'alice@test.com');
      expect(found.id).to.equal(user.id);
    });

    it('should return null when the user does not exist', async () => {
      const found = await userRepo.findById(99999);

      expect(found).to.be.null;
    });
  });

  describe('#findByEmail', () => {
    it('should perform a case-insensitive email lookup', async () => {
      await userRepo.create({ name: 'Bob', email: 'Bob@Example.COM' });
      const found = await userRepo.findByEmail('bob@example.com');

      expect(found).to.not.be.null;
      expect(found.name).to.equal('Bob');
    });
  });

  context('when creating users with constraints', () => {
    it('should enforce the unique email constraint', async () => {
      await userRepo.create({ name: 'First', email: 'unique@test.com' });

      try {
        await userRepo.create({ name: 'Second', email: 'unique@test.com' });
        expect.fail('Expected a constraint violation error');
      } catch (err) {
        expect(err.message).to.include('UNIQUE');
      }
    });

    it('should set createdAt and updatedAt timestamps automatically', async () => {
      const user = await userRepo.create({ name: 'Timestamp', email: 'ts@test.com' });

      expect(user.createdAt).to.be.an.instanceOf(Date);
      expect(user.updatedAt).to.be.an.instanceOf(Date);
    });

    it('should trim whitespace from the name field', async () => {
      const user = await userRepo.create({ name: '  Padded  ', email: 'pad@test.com' });

      expect(user.name).to.equal('Padded');
    });
  });

  context('when querying with associations', () => {
    it('should include the order count via a subquery', async () => {
      const user = await userRepo.create({ name: 'Shopper', email: 'shop@test.com' });
      await orderRepo.create({ userId: user.id, total: 29.99 });
      await orderRepo.create({ userId: user.id, total: 49.99 });

      const result = await userRepo.findWithOrderCount(user.id);

      expect(result.orderCount).to.equal(2);
    });

    it('should return zero order count for users with no orders', async () => {
      const user = await userRepo.create({ name: 'Lurker', email: 'lurk@test.com' });

      const result = await userRepo.findWithOrderCount(user.id);

      expect(result.orderCount).to.equal(0);
    });
  });

  describe('#bulkInsert', () => {
    it('should insert multiple users in a single transaction', async () => {
      const users = [
        { name: 'Batch1', email: 'b1@test.com' },
        { name: 'Batch2', email: 'b2@test.com' },
        { name: 'Batch3', email: 'b3@test.com' },
      ];

      const inserted = await userRepo.bulkInsert(users);

      expect(inserted).to.have.lengthOf(3);
      expect(inserted[0]).to.have.property('id');
    });
  });
});

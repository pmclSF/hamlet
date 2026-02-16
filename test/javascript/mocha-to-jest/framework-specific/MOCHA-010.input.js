const { expect } = require('chai');
const sinon = require('sinon');

describe('UserService', () => {
  let userService;

  before(() => {
    // Initialize service
  });

  after(() => {
    // Cleanup
  });

  beforeEach(() => {
    userService = { getUser: sinon.stub() };
  });

  afterEach(() => {
    sinon.restore();
  });

  context('when user exists', () => {
    it('returns the user', async () => {
      userService.getUser.returns({ name: 'Alice' });
      const user = userService.getUser();
      expect(user).to.deep.equal({ name: 'Alice' });
      sinon.assert.calledOnce(userService.getUser);
    });

    specify('user has a name', () => {
      userService.getUser.returns({ name: 'Bob' });
      const user = userService.getUser();
      expect(user.name).to.be.a('string');
      expect(user.name).to.have.lengthOf(3);
    });
  });
});

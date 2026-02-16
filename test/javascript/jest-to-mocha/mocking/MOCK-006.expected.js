const { expect } = require('chai');
const sinon = require('sinon');

describe('test', () => {
  it('rejects', async () => {
    const fn = sinon.stub().rejects(new Error('fail'));
    try {
      await fn();
    } catch (e) {
      expect(e.message).to.equal('fail');
    }
  });
});

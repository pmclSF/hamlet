const { expect } = require('chai');
const sinon = require('sinon');
describe('test', () => {
  it('returns', () => {
    const fn = sinon.stub().returns(42);
    expect(fn()).to.equal(42);
  });
});

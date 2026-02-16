const { expect } = require('chai');
const sinon = require('sinon');

describe('test', () => {
  it('called times', () => {
    const fn = sinon.stub();
    fn();
    fn();
    expect(fn.callCount).to.equal(2);
  });
});

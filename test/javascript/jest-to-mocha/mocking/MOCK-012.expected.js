const { expect } = require('chai');
const sinon = require('sinon');

describe('test', () => {
  it('clear', () => {
    const fn = sinon.stub();
    fn();
    fn.resetHistory();
    expect(fn).to.not.have.been.called;
  });
});

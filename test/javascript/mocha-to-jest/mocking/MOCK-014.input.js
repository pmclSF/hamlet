const { expect } = require('chai');
const sinon = require('sinon');
describe('test', () => {
  it('chai-sinon calledWith', () => {
    const fn = sinon.stub();
    fn('x');
    expect(fn).to.have.been.calledWith('x');
  });
});

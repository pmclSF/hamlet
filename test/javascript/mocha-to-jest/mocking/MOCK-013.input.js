const { expect } = require('chai');
const sinon = require('sinon');
describe('test', () => {
  it('chai-sinon', () => {
    const fn = sinon.stub();
    fn();
    expect(fn).to.have.been.calledOnce;
  });
});

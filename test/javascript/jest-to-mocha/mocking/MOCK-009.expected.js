const { expect } = require('chai');
const sinon = require('sinon');

describe('test', () => {
  it('not called', () => {
    const fn = sinon.stub();
    expect(fn).to.not.have.been.called;
  });
});

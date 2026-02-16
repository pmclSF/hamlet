const { expect } = require('chai');
const sinon = require('sinon');

describe('test', () => {
  it('mock', () => {
    const fn = sinon.stub();
    fn();
    expect(fn).to.have.been.called;
  });
});

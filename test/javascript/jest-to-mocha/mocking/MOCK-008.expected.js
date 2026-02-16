const { expect } = require('chai');
const sinon = require('sinon');

describe('test', () => {
  it('called with', () => {
    const fn = sinon.stub();
    fn('a', 'b');
    expect(fn).to.have.been.calledWith('a', 'b');
  });
});

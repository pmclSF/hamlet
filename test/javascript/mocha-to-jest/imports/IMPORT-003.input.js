const { expect } = require('chai');
const sinon = require('sinon');

describe('test', () => {
  it('works', () => {
    const fn = sinon.stub();
    fn();
    expect(fn()).to.be.undefined;
  });
});

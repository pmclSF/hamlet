const { expect } = require('chai');
const sinon = require('sinon');

describe('test', () => {
  it('impl', () => {
    const fn = sinon.stub().callsFake(x => x * 2);
    expect(fn(5)).to.equal(10);
  });
});

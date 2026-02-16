const { expect } = require('chai');
const sinon = require('sinon');

describe('test', () => {
  it('resolves', async () => {
    const fn = sinon.stub().resolves('data');
    const result = await fn();
    expect(result).to.equal('data');
  });
});

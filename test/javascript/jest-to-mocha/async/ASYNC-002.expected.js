const { expect } = require('chai');

describe('test', () => {
  it('async', async () => {
    const val = await Promise.resolve(42);
    expect(val).to.equal(42);
  });
});

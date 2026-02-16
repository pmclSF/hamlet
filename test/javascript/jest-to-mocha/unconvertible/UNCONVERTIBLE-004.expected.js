const { expect } = require('chai');

describe('test', () => {
  it('inline snapshot', () => {
    expect(1).to.equal(1);
    // HAMLET-TODO [UNCONVERTIBLE-INLINE-SNAPSHOT]: Mocha does not support inline snapshots
// Original: expect('hello').toMatchInlineSnapshot('hello');
// Manual action required: Convert to explicit assertion
// expect('hello').toMatchInlineSnapshot('hello');
  });
});

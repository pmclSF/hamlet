const { expect } = require('chai');

describe('test', () => {
  it('snapshot', () => {
    expect(1).to.equal(1);
    // HAMLET-TODO [UNCONVERTIBLE-SNAPSHOT]: Mocha does not have built-in snapshot testing
// Original: expect({ a: 1 }).toMatchSnapshot();
// Manual action required: Use chai-jest-snapshot or snap-shot-it package
// expect({ a: 1 }).toMatchSnapshot();
  });
});

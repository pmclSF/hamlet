const { expect } = require('chai');
const sinon = require('sinon');

// HAMLET-TODO [UNCONVERTIBLE-MODULE-MOCK]: Mocha does not have a built-in module mocking system like jest.mock()
// Original: jest.mock('./api');
// Manual action required: Use proxyquire, rewire, or manual dependency injection
// jest.mock('./api');

describe('test', () => {
  it('full', () => {
    const fn = sinon.stub().returns(42);
    const result = fn();
    expect(result).to.equal(42);
    expect(fn.callCount).to.equal(1);
  });
});

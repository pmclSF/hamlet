const { expect } = require('chai');
const sinon = require('sinon');

// HAMLET-TODO [UNCONVERTIBLE-MODULE-MOCK]: Mocha does not have a built-in module mocking system like jest.mock()
// Original: jest.mock('./utils');
// Manual action required: Use proxyquire, rewire, or manual dependency injection
// jest.mock('./utils');

describe('test', () => {
  it('works', () => {
    const fn = sinon.stub();
    fn();
    expect(fn).to.have.been.called;
  });
});

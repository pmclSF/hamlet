const { expect } = require('chai');
describe('test', () => {
  it('include', () => {
    expect([1, 2, 3]).to.include(2);
  });
});

const { expect } = require('chai');
describe('test', () => {
  it('length', () => {
    expect([1, 2, 3]).to.have.lengthOf(3);
  });
});

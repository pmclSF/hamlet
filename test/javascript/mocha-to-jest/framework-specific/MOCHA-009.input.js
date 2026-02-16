const { expect } = require('chai');
describe('test', () => {
  it('multiple', () => {
    expect(1).to.equal(1);
    expect('hello').to.have.lengthOf(5);
    expect([1, 2]).to.include(1);
    expect({ a: 1 }).to.have.property('a');
  });
});

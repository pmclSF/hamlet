const { expect } = require('chai');
describe('test', () => {
  it('property', () => {
    expect({ a: 1 }).to.have.property('a');
  });
});

const { expect } = require('chai');
describe('test', () => {
  it('close to', () => {
    expect(0.1 + 0.2).to.be.closeTo(0.3, 0.01);
  });
});

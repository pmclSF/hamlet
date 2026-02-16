const { expect } = require('chai');
describe('test', () => {
  it('instanceOf', () => {
    expect(new Date()).to.be.an.instanceOf(Date);
  });
});

const { expect } = require('chai');
describe('test', function() {
  this.retries(3);
  it('works', () => {
    expect(true).to.be.true;
  });
});

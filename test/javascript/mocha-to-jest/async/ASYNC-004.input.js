const { expect } = require('chai');
describe('test', function() {
  this.timeout(5000);

  it('slow test', () => {
    expect(true).to.be.true;
  });
});

const { expect } = require('chai');

describe('test', () => {
  it('async', (done) => {
    setTimeout(() => {
      expect(true).to.be.true;
      done();
    }, 100);
  });
});

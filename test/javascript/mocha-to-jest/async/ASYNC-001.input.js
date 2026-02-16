const { expect } = require('chai');
describe('test', () => {
  it('async with done', (done) => {
    setTimeout(() => {
      expect(true).to.be.true;
      done();
    }, 100);
  });
});

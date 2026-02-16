const { expect } = require('chai');
describe('test', () => {
  it('throw', () => {
    expect(throwingFn).to.throw();
  });
});

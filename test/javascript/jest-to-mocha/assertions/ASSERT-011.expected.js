const { expect } = require('chai');

describe('test', () => {
  it('less', () => {
    expect(3).to.be.below(10);
  });
});

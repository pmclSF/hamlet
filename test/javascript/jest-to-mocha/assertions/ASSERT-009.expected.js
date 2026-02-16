const { expect } = require('chai');

describe('test', () => {
  it('falsy', () => {
    expect(0).to.not.be.ok;
  });
});

const { expect } = require('chai');

describe('test', () => {
  it('contain', () => {
    expect([1, 2, 3]).to.include(2);
  });
});

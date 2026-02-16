const { expect } = require('chai');

describe('test', () => {
  it('throws', () => {
    expect(() => { throw new Error('fail'); }).to.throw();
  });
});

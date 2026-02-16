const { expect } = require('chai');

describe('test', () => {
  it('deep equal', () => {
    expect({ a: 1 }).to.deep.equal({ a: 1 });
  });
});

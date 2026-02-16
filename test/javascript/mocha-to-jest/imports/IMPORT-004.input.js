const { expect } = require('chai');
const path = require('path');

describe('test', () => {
  it('works', () => {
    expect(path.join('a', 'b')).to.equal('a/b');
  });
});

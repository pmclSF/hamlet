const { expect } = require('chai');
describe('test', () => {
  it('match', () => {
    expect('hello world').to.match(/hello/);
  });
});

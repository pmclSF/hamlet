const { assert } = require('chai');
describe('test', () => {
  it('assert deep', () => {
    assert.deepEqual({ a: 1 }, { a: 1 });
  });
});

const { assert } = require('chai');
describe('test', () => {
  it('include and match', () => {
    assert.include('hello world', 'hello');
    assert.match('hello', /hel/);
  });
});

const sinon = require('sinon');
describe('test', () => {
  it('calledWith', () => {
    const fn = sinon.stub();
    fn('a', 'b');
    sinon.assert.calledWith(fn, 'a', 'b');
  });
});

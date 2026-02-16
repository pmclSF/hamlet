const sinon = require('sinon');
describe('test', () => {
  it('callCount', () => {
    const fn = sinon.stub();
    fn();
    fn();
    fn();
    sinon.assert.callCount(fn, 3);
  });
});

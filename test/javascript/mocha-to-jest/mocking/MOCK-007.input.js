const sinon = require('sinon');
describe('test', () => {
  it('calledOnce', () => {
    const fn = sinon.stub();
    fn();
    sinon.assert.calledOnce(fn);
  });
});

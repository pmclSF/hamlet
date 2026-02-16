const sinon = require('sinon');
describe('test', () => {
  it('notCalled', () => {
    const fn = sinon.stub();
    sinon.assert.notCalled(fn);
  });
});

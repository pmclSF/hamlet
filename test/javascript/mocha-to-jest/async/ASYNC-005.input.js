const sinon = require('sinon');
const { expect } = require('chai');
describe('test', () => {
  it('fake timers', () => {
    sinon.useFakeTimers();
    const fn = sinon.stub();
    setTimeout(fn, 1000);
    clock.tick(1000);
    sinon.assert.calledOnce(fn);
    clock.restore();
  });
});

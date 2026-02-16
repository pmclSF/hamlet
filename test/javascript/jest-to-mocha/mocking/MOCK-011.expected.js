const { expect } = require('chai');
const sinon = require('sinon');

describe('test', () => {
  it('timers', () => {
    sinon.useFakeTimers();
    const fn = sinon.stub();
    setTimeout(fn, 1000);
    clock.tick(1000);
    expect(fn).to.have.been.called;
    clock.restore();
  });
});

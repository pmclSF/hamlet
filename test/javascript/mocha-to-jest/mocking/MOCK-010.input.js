const sinon = require('sinon');
describe('test', () => {
  it('timers', () => {
    sinon.useFakeTimers();
    setTimeout(() => {}, 1000);
    clock.tick(1000);
    clock.restore();
  });
});

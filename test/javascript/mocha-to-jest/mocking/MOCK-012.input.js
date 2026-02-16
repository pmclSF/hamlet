const sinon = require('sinon');
const sandbox = sinon.createSandbox();
describe('test', () => {
  afterEach(() => {
    sandbox.restore();
  });

  it('works', () => {
    const fn = sinon.stub();
    fn();
  });
});

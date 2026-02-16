const sinon = require('sinon');

describe('test', () => {
  afterEach(() => {
    sinon.restore();
  });

  it('works', () => {
    const fn = sinon.stub();
    fn();
  });
});

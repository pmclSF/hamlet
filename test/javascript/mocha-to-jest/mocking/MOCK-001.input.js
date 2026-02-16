const { expect } = require('chai');
const sinon = require('sinon');
describe('test', () => {
  it('stub', () => {
    const fn = sinon.stub();
    fn();
    sinon.assert.called(fn);
  });
});

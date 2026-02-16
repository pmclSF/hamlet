const { expect } = require('chai');
const sinon = require('sinon');

describe('test', () => {
  it('spy', () => {
    const obj = { foo: () => 42 };
    sinon.spy(obj, 'foo');
    obj.foo();
    expect(obj.foo).to.have.been.called;
  });
});

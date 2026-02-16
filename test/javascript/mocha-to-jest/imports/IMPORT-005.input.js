import { expect } from 'chai';
import sinon from 'sinon';

describe('test', () => {
  it('works', () => {
    const fn = sinon.stub();
    expect(fn()).to.be.undefined;
  });
});

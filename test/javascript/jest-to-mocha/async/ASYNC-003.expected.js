const { expect } = require('chai');

describe('test', () => {
  it('promise', () => {
    return Promise.resolve(42).then(val => {
      expect(val).to.equal(42);
    });
  });
});

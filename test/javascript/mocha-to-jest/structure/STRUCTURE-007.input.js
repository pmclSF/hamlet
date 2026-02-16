const { expect } = require('chai');

describe('App', () => {
  before(() => {
    // init
  });

  context('feature A', () => {
    specify('works', () => {
      expect('a').to.equal('a');
    });
  });

  after(() => {
    // cleanup
  });
});

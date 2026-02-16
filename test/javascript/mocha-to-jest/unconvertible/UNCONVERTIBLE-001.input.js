const chai = require('chai');
const chaiAsPromised = require('chai-as-promised');
chai.use(chaiAsPromised);

describe('test', () => {
  it('works', () => {
    chai.expect(true).to.be.true;
  });
});

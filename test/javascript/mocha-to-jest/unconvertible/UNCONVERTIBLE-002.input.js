const sinon = require('sinon');
describe('test', () => {
  it('server', () => {
    sinon.fakeServer.create();
  });
});

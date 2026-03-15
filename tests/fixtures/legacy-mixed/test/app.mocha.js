import { expect } from 'chai';
import { startApp, stopApp, getStatus } from '../src/app.js';

describe('App (Mocha)', function() {
  it('should start app on default port', function() {
    const app = startApp();
    expect(app.running).to.equal(true);
    expect(app.port).to.equal(3000);
  });

  it('should stop app', function() {
    const app = startApp();
    const stopped = stopApp(app);
    expect(stopped.running).to.equal(false);
  });

  it('should report status', function() {
    const app = startApp();
    expect(getStatus(app)).to.equal('healthy');
  });
});

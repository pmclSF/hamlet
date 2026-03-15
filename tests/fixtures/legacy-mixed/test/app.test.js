import { startApp, stopApp, getStatus } from '../src/app.js';

describe('App (Jest)', () => {
  test('should start app on default port', () => {
    const app = startApp();
    expect(app.running).toBe(true);
    expect(app.port).toBe(3000);
  });

  test('should stop app', () => {
    const app = startApp();
    const stopped = stopApp(app);
    expect(stopped.running).toBe(false);
  });

  test('should report status', () => {
    const app = startApp();
    expect(getStatus(app)).toBe('healthy');
    expect(getStatus(stopApp(app))).toBe('stopped');
  });
});

export function startApp(port) {
  return { running: true, port: port || 3000 };
}

export function stopApp(app) {
  return { ...app, running: false };
}

export function getStatus(app) {
  return app.running ? 'healthy' : 'stopped';
}

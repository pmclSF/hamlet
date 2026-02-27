import path from 'path';
import { fileURLToPath } from 'url';
import { HamletServer } from '../../src/server/HamletServer.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FIXTURES_ANALYZE = path.resolve(__dirname, '../fixtures/analyze');

describe('HamletServer UI mode', () => {
  let server;
  let baseUrl;

  beforeAll(async () => {
    server = new HamletServer({ port: 0, root: FIXTURES_ANALYZE, serveUI: true });
    baseUrl = await server.start();
  });

  afterAll(async () => {
    await server.stop();
  });

  it('should serve /api/health with root in response', async () => {
    const res = await fetch(`${baseUrl}/api/health`);
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.status).toBe('ok');
    expect(body.root).toBeDefined();
    expect(body.root).toContain('fixtures');
  });

  it('should serve index.html at /', async () => {
    const res = await fetch(`${baseUrl}/`);
    expect(res.status).toBe(200);
    expect(res.headers.get('content-type')).toContain('text/html');
    const text = await res.text();
    expect(text).toContain('Hamlet');
    expect(text).toContain('<script');
  });

  it('should serve styles.css', async () => {
    const res = await fetch(`${baseUrl}/styles.css`);
    expect(res.status).toBe(200);
    expect(res.headers.get('content-type')).toContain('text/css');
  });

  it('should serve app.js', async () => {
    const res = await fetch(`${baseUrl}/app.js`);
    expect(res.status).toBe(200);
    expect(res.headers.get('content-type')).toContain('javascript');
  });

  it('should SPA-fallback unknown paths to index.html', async () => {
    const res = await fetch(`${baseUrl}/analyze`);
    expect(res.status).toBe(200);
    expect(res.headers.get('content-type')).toContain('text/html');
    const text = await res.text();
    expect(text).toContain('Hamlet');
  });

  it('should still return JSON 404 for unknown /api/ routes', async () => {
    const res = await fetch(`${baseUrl}/api/nonexistent`);
    expect(res.status).toBe(404);
    const body = await res.json();
    expect(body.error).toBeDefined();
  });

  it('should serve /api/file endpoint', async () => {
    const testFile = path.resolve(FIXTURES_ANALYZE, 'auth.test.js');
    const res = await fetch(
      `${baseUrl}/api/file?path=${encodeURIComponent(testFile)}`
    );
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.content).toContain('describe');
  });

  it('should return 400 for /api/file without path', async () => {
    const res = await fetch(`${baseUrl}/api/file`);
    expect(res.status).toBe(400);
  });

  it('should return 404 for /api/file with nonexistent path', async () => {
    const res = await fetch(
      `${baseUrl}/api/file?path=${encodeURIComponent('/nonexistent/file.js')}`
    );
    expect(res.status).toBe(404);
  });
});

import http from 'node:http';
import fsSync from 'node:fs';
import nodePath from 'node:path';
import { fileURLToPath } from 'node:url';
import { Router, sendJson } from './router.js';
import {
  handleHealth,
  handleAnalyze,
  handleConvert,
  handleGetJob,
  handleJobStream,
  handleArtifacts,
  handleOpen,
  handleFile,
  handlePreview,
} from './handlers.js';

const __dirname = nodePath.dirname(fileURLToPath(import.meta.url));
const UI_DIR = nodePath.join(__dirname, '..', 'ui');

const ALLOWED_HOSTS = new Set(['localhost', '127.0.0.1', '[::1]']);

const MIME_TYPES = {
  '.html': 'text/html; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.js': 'application/javascript; charset=utf-8',
  '.json': 'application/json',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.ico': 'image/x-icon',
};

function isLocalhostHost(header) {
  if (!header) return false;
  // Strip port suffix if present
  const host = header.replace(/:\d+$/, '');
  return ALLOWED_HOSTS.has(host);
}

export class HamletServer {
  /**
   * @param {Object} [options]
   * @param {number} [options.port=0] - Port to bind (0 = random)
   * @param {string} [options.root='.'] - Project root directory
   * @param {boolean} [options.serveUI=false] - Serve browser UI static files
   * @param {boolean} [options.enableOpen=false] - Enable /api/open endpoint.
   *   Disabled by default because it invokes OS-level open/xdg-open on
   *   user-supplied paths, which could be abused by any local network caller
   *   to launch arbitrary applications or URLs. When enabled, paths are
   *   restricted to the project root.
   */
  constructor({
    port = 0,
    root = '.',
    serveUI = false,
    enableOpen = false,
  } = {}) {
    this._port = port;
    this._root = root;
    this._serveUI = serveUI;
    this._enableOpen = enableOpen;
    this._server = null;
  }

  /**
   * Start the server.
   * @returns {Promise<string>} URL the server is listening on
   */
  start() {
    const resolvedRoot = nodePath.resolve(this._root);

    return new Promise((resolve, reject) => {
      const router = new Router();

      // Register routes
      router.get('/api/health', handleHealth);
      router.post('/api/analyze', handleAnalyze);
      router.post('/api/convert', handleConvert);
      router.get('/api/jobs/:id', handleGetJob);
      router.get('/api/jobs/:id/stream', handleJobStream);
      router.get('/api/artifacts/:jobId', handleArtifacts);
      if (this._enableOpen) {
        router.post('/api/open', handleOpen);
      }
      router.post('/api/preview', handlePreview);
      router.get('/api/file', handleFile);

      this._server = http.createServer(async (req, res) => {
        // Host header validation — localhost only
        if (!isLocalhostHost(req.headers.host)) {
          sendJson(res, 403, { error: 'Forbidden: non-localhost request' });
          return;
        }

        // Attach project root for handlers
        req.serverRoot = resolvedRoot;

        try {
          const matched = await router.dispatch(req, res);
          if (!matched) {
            const pathname = new URL(req.url, `http://${req.headers.host}`)
              .pathname;
            if (this._serveUI && !pathname.startsWith('/api/')) {
              this._serveStatic(req, res);
            } else {
              sendJson(res, 404, { error: 'Not found' });
            }
          }
        } catch (err) {
          if (err.statusCode) {
            sendJson(res, err.statusCode, { error: err.message });
          } else if (err instanceof SyntaxError) {
            sendJson(res, 400, { error: err.message });
          } else {
            sendJson(res, 500, { error: err.message });
          }
        }
      });

      this._server.listen(this._port, '127.0.0.1', () => {
        const addr = this._server.address();
        resolve(`http://127.0.0.1:${addr.port}`);
      });

      this._server.on('error', reject);
    });
  }

  /**
   * Serve static files from the UI directory, with SPA fallback.
   */
  _serveStatic(req, res) {
    const url = new URL(req.url, `http://${req.headers.host}`);
    let filePath = nodePath.join(UI_DIR, url.pathname);

    // Prevent path traversal
    if (!filePath.startsWith(UI_DIR)) {
      sendJson(res, 403, { error: 'Forbidden' });
      return;
    }

    // Check if file exists
    try {
      const stat = fsSync.statSync(filePath);
      if (stat.isFile()) {
        const ext = nodePath.extname(filePath);
        const mime = MIME_TYPES[ext] || 'application/octet-stream';
        const content = fsSync.readFileSync(filePath);
        res.writeHead(200, { 'Content-Type': mime });
        res.end(content);
        return;
      }
    } catch (_e) {
      // File not found — fall through to SPA fallback
    }

    // SPA fallback: serve index.html
    try {
      const indexPath = nodePath.join(UI_DIR, 'index.html');
      const content = fsSync.readFileSync(indexPath);
      res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
      res.end(content);
    } catch (_e) {
      sendJson(res, 404, { error: 'UI files not found' });
    }
  }

  /**
   * The port the server is listening on, or null if not started.
   * @returns {number|null}
   */
  get address() {
    const addr = this._server && this._server.address();
    return addr ? addr.port : null;
  }

  /**
   * Stop the server gracefully.
   * @returns {Promise<void>}
   */
  stop() {
    return new Promise((resolve, reject) => {
      if (!this._server) return resolve();
      this._server.closeAllConnections();
      this._server.close((err) => {
        if (err) return reject(err);
        this._server = null;
        resolve();
      });
    });
  }
}

import http from 'node:http';
import { Router, sendJson } from './router.js';
import {
  handleHealth,
  handleAnalyze,
  handleConvert,
  handleGetJob,
  handleJobStream,
  handleArtifacts,
  handleOpen,
} from './handlers.js';

const ALLOWED_HOSTS = new Set(['localhost', '127.0.0.1', '[::1]']);

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
   */
  constructor({ port = 0, root = '.' } = {}) {
    this._port = port;
    this._root = root;
    this._server = null;
  }

  /**
   * Start the server.
   * @returns {Promise<string>} URL the server is listening on
   */
  start() {
    return new Promise((resolve, reject) => {
      const router = new Router();

      // Register routes
      router.get('/api/health', handleHealth);
      router.post('/api/analyze', handleAnalyze);
      router.post('/api/convert', handleConvert);
      router.get('/api/jobs/:id', handleGetJob);
      router.get('/api/jobs/:id/stream', handleJobStream);
      router.get('/api/artifacts/:jobId', handleArtifacts);
      router.post('/api/open', handleOpen);

      this._server = http.createServer(async (req, res) => {
        // Host header validation — localhost only
        if (!isLocalhostHost(req.headers.host)) {
          sendJson(res, 403, { error: 'Forbidden: non-localhost request' });
          return;
        }

        try {
          const matched = await router.dispatch(req, res);
          if (!matched) {
            sendJson(res, 404, { error: 'Not found' });
          }
        } catch (err) {
          if (err instanceof SyntaxError) {
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

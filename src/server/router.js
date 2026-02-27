/**
 * Send a JSON response.
 * @param {import('http').ServerResponse} res
 * @param {number} status
 * @param {*} data
 */
export function sendJson(res, status, data) {
  const body = JSON.stringify(data);
  res.writeHead(status, {
    'Content-Type': 'application/json',
    'Content-Length': Buffer.byteLength(body),
  });
  res.end(body);
}

export class Router {
  constructor() {
    /** @type {Array<{ method: string, regex: RegExp, keys: string[], handler: Function }>} */
    this._routes = [];
  }

  /**
   * Register a GET route.
   * @param {string} pattern - e.g. "/api/jobs/:id"
   * @param {Function} handler - (req, res) => void
   */
  get(pattern, handler) {
    this._add('GET', pattern, handler);
  }

  /**
   * Register a POST route.
   * @param {string} pattern - e.g. "/api/convert"
   * @param {Function} handler - (req, res) => void
   */
  post(pattern, handler) {
    this._add('POST', pattern, handler);
  }

  _add(method, pattern, handler) {
    const keys = [];
    const regexStr = pattern.replace(/:([^/]+)/g, (_match, key) => {
      keys.push(key);
      return '([^/]+)';
    });
    this._routes.push({
      method,
      regex: new RegExp(`^${regexStr}$`),
      keys,
      handler,
    });
  }

  /**
   * Dispatch a request to the matching route.
   * @param {import('http').IncomingMessage} req
   * @param {import('http').ServerResponse} res
   * @returns {Promise<boolean>} true if a route matched
   */
  async dispatch(req, res) {
    const url = new URL(req.url, `http://${req.headers.host || 'localhost'}`);
    const pathname = url.pathname;

    for (const route of this._routes) {
      if (route.method !== req.method) continue;
      const match = pathname.match(route.regex);
      if (!match) continue;

      req.params = {};
      route.keys.forEach((key, i) => {
        req.params[key] = decodeURIComponent(match[i + 1]);
      });

      if (req.method === 'POST') {
        req.body = await readJsonBody(req);
      }

      await route.handler(req, res);
      return true;
    }

    return false;
  }
}

/** Maximum request body size in bytes (1 MB). */
const MAX_BODY_SIZE = 1 * 1024 * 1024;

/**
 * Custom error for payloads that exceed the size limit.
 * Handlers can check `err.statusCode` to return the correct HTTP status.
 */
class PayloadTooLargeError extends Error {
  constructor() {
    super('Request body exceeds maximum allowed size');
    this.statusCode = 413;
  }
}

/**
 * Read and parse JSON from request body, enforcing a size limit.
 *
 * On overflow, listeners are removed deterministically and the request stream
 * is destroyed so that abusive senders cannot keep writing data.
 *
 * @param {import('http').IncomingMessage} req
 * @returns {Promise<Object>}
 */
function readJsonBody(req) {
  return new Promise((resolve, reject) => {
    // Reject early if Content-Length header already exceeds the limit
    const contentLength = parseInt(req.headers['content-length'], 10);
    if (contentLength > MAX_BODY_SIZE) {
      req.resume();
      return reject(new PayloadTooLargeError());
    }

    const chunks = [];
    let received = 0;
    let settled = false;

    function cleanup() {
      req.removeListener('data', onData);
      req.removeListener('end', onEnd);
      req.removeListener('error', onError);
    }

    function onData(chunk) {
      received += chunk.length;
      if (received > MAX_BODY_SIZE) {
        settled = true;
        cleanup();
        chunks.length = 0;
        // Destroy the stream to stop receiving data from abusive senders.
        // The 413 response is written by the error handler in HamletServer
        // before the socket is torn down because node:http flushes the
        // response before closing.
        req.destroy();
        return reject(new PayloadTooLargeError());
      }
      chunks.push(chunk);
    }

    function onEnd() {
      cleanup();
      const raw = Buffer.concat(chunks).toString();
      if (!raw) return resolve({});
      try {
        resolve(JSON.parse(raw));
      } catch (_err) {
        reject(new SyntaxError('Invalid JSON body'));
      }
    }

    function onError(err) {
      if (!settled) {
        cleanup();
        reject(err);
      }
    }

    req.on('data', onData);
    req.on('end', onEnd);
    req.on('error', onError);
  });
}

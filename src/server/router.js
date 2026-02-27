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

/**
 * Read and parse JSON from request body.
 * @param {import('http').IncomingMessage} req
 * @returns {Promise<Object>}
 */
function readJsonBody(req) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    req.on('data', (chunk) => chunks.push(chunk));
    req.on('end', () => {
      const raw = Buffer.concat(chunks).toString();
      if (!raw) return resolve({});
      try {
        resolve(JSON.parse(raw));
      } catch (err) {
        reject(new SyntaxError('Invalid JSON body'));
      }
    });
    req.on('error', reject);
  });
}

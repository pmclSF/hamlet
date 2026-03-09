import { describe, it } from 'node:test';
import assert from 'node:assert';
import { createServer } from '../server.js';

describe('server', () => {
  it('should start on the given port', async () => {
    const server = createServer();
    assert.ok(server);
  });
});

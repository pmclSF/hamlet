import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('CONTEXT-003: Test context via closures', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'CONTEXT-003');
  });
});

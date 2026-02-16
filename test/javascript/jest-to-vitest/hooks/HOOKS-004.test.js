import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('HOOKS-004: afterEach per-test teardown', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'HOOKS-004');
  });
});

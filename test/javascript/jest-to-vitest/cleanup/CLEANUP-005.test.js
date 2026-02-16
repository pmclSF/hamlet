import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('CLEANUP-005: Context manager equivalent with try/finally', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'CLEANUP-005');
  });
});

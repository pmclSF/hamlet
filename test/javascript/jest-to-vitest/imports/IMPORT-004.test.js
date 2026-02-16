import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('IMPORT-004: Relative imports (test helpers)', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'IMPORT-004');
  });
});

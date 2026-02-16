import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('CLEANUP-006: try/finally for cleanup with multiple resources', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'CLEANUP-006');
  });
});

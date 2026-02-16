import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('CLEANUP-002: Teardown runs even when setup fails', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'CLEANUP-002');
  });
});

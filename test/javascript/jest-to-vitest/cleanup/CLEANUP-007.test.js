import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('CLEANUP-007: Temporary file cleanup pattern', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'CLEANUP-007');
  });
});

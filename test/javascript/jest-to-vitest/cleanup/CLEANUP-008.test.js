import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('CLEANUP-008: afterAll for shared resource cleanup', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'CLEANUP-008');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('TS-011: Strict null checks patterns', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'TS-011', { ext: '.ts' });
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('UNCONV-004: test generators pass through unchanged', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'UNCONV-004');
  });
});

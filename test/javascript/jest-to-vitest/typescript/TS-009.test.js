import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('TS-009: satisfies operator', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'TS-009', { ext: '.ts' });
  });
});

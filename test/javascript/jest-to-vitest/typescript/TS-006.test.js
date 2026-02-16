import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('TS-006: Typed mock return values', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'TS-006', { ext: '.ts' });
  });
});

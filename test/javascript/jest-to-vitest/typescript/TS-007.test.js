import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('TS-007: Enum usage in test data', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'TS-007', { ext: '.ts' });
  });
});

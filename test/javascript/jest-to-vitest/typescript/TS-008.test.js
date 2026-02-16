import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('TS-008: as const assertions', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'TS-008', { ext: '.ts' });
  });
});

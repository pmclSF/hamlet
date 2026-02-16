import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('TS-004: Type-only imports (import type)', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'TS-004', { ext: '.ts' });
  });
});

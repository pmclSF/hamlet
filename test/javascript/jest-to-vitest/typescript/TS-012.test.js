import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('TS-012: Runtime type checking with typeof', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'TS-012', { ext: '.ts' });
  });
});

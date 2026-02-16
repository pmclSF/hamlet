import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('TS-001: Type annotations on test variables', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'TS-001', { ext: '.ts' });
  });
});

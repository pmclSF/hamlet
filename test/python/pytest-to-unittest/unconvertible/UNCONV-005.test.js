import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('UNCONV-005: conftest fixture reference passes through', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'UNCONV-005');
  });
});

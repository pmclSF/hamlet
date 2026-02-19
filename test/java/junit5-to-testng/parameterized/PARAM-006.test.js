import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('PARAM-006: Import cleanup for parameterized imports', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'PARAM-006');
  });
});

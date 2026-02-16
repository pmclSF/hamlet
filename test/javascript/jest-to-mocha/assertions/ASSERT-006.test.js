import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ASSERT-006: toBeUndefined to to.be.undefined', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ASSERT-006');
  });
});

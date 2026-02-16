import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ASSERT-004: assertNull and assertNotNull (simple rename)', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ASSERT-004');
  });
});

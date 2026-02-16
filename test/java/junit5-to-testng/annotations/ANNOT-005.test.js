import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ANNOT-005: @Disabled → @Test(enabled = false)', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ANNOT-005');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MARKER-005: Custom marker preserved as-is', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MARKER-005');
  });
});

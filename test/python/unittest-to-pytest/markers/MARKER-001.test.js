import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MARKER-001: unittest.skip to pytest.mark.skip', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MARKER-001');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MARKER-004: pytest.mark.xfail to unittest.expectedFailure', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MARKER-004');
  });
});

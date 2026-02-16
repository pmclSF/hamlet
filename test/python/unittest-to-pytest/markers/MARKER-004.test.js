import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MARKER-004: unittest.expectedFailure to pytest.mark.xfail', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MARKER-004');
  });
});

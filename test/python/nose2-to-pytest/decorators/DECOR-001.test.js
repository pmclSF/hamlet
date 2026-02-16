import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('DECOR-001: @params(...) to @pytest.mark.parametrize', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'DECOR-001');
  });
});

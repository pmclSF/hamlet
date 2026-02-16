import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('DECOR-002: @attr(tag) to @pytest.mark.tag', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'DECOR-002');
  });
});

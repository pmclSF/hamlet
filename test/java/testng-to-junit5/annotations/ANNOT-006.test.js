import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ANNOT-006: @Test(groups = {"x"}) → @Tag("x") + @Test', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ANNOT-006');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('FIXTURE-005: tearDown with multi-line cleanup', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'FIXTURE-005');
  });
});

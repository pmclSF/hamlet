import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('FWSPEC-002: Full realistic conversion', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'FWSPEC-002');
  });
});

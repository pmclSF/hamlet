import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('FIXTURE-006: Multiple fixtures mixed conversion', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'FIXTURE-006');
  });
});

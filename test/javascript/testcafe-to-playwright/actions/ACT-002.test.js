import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ACT-002: t.click', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ACT-002');
  });
});

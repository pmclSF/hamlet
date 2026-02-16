import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('VALUE-009: Large numbers / BigInt', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'VALUE-009');
  });
});

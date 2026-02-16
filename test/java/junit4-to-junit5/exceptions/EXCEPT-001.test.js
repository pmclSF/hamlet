import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('EXCEPT-001: @Test(expected = X.class) → assertThrows', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'EXCEPT-001');
  });
});

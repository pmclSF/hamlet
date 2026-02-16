import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('VALUE-015: Deep equality with nested objects', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'VALUE-015');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('IMPORT-003: multiple nose imports removed, pytest added when needed', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'IMPORT-003');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('IMPORT-001: Adds chai import when assertions used', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'IMPORT-001');
  });
});

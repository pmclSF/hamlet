import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('VALUE-001: Template literals in assertions', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'VALUE-001');
  });
});

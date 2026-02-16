import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('VALUE-005: Special characters in selectors/strings', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'VALUE-005');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('VALUE-003: Unicode characters in test names and assertions', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'VALUE-003');
  });
});

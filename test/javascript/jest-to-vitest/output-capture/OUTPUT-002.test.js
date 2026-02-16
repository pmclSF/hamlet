import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('OUTPUT-002: Assert on stderr via console.error spy', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'OUTPUT-002');
  });
});

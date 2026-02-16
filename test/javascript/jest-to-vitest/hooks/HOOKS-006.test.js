import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('HOOKS-006: Multiple hooks of same type', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'HOOKS-006');
  });
});

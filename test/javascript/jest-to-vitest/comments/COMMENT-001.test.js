import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('COMMENT-001: Inline comments within test body', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'COMMENT-001');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('COMMENT-005: Comment explaining assertion', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'COMMENT-005');
  });
});

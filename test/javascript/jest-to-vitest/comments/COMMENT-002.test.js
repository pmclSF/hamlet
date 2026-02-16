import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('COMMENT-002: Block comments above test', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'COMMENT-002');
  });
});

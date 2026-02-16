import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('COMMENT-008: License header', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'COMMENT-008');
  });
});

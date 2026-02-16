import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('COMMENT-007: Framework-specific directives (eslint-disable)', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'COMMENT-007');
  });
});

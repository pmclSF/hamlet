import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ASSERT-015: not.toBe to to.not.equal', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ASSERT-015');
  });
});

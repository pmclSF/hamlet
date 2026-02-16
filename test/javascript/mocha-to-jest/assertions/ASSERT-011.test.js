import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ASSERT-011: to.include to toContain', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ASSERT-011');
  });
});

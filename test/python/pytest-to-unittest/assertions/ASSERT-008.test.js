import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ASSERT-008: assert a not in b to assertNotIn', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ASSERT-008');
  });
});

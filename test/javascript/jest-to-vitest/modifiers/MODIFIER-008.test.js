import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MODIFIER-008: Test timeout override', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MODIFIER-008');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('PARAM-008: Parameterized async tests', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'PARAM-008');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('HOOKS-008: Hooks with timeout parameter', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'HOOKS-008');
  });
});

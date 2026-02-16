import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('HOOKS-010: Hooks sharing state via closure variables', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'HOOKS-010');
  });
});

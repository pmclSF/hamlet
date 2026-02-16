import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('HOOKS-005: Nested hooks with outer and inner beforeEach', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'HOOKS-005');
  });
});

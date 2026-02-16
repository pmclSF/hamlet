import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MODIFIER-005: Conditional skip with reason', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MODIFIER-005');
  });
});

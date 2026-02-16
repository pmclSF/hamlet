import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('STRUCTURE-005: Mixed hooks', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'STRUCTURE-005');
  });
});

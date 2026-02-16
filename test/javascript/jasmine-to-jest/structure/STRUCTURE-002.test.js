import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('STRUCTURE-002: Nested describe', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'STRUCTURE-002');
  });
});

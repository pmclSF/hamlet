import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('STRUCTURE-007: Mixed context, specify, before, after', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'STRUCTURE-007');
  });
});

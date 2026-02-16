import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ANNOT-007: Combined annotations', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ANNOT-007');
  });
});

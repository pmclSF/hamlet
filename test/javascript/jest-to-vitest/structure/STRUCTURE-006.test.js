import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('STRUCTURE-006: Empty test bodies', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'STRUCTURE-006');
  });
});

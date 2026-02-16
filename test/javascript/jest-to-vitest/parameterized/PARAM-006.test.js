import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('PARAM-006: Dynamic test generation from array', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'PARAM-006');
  });
});

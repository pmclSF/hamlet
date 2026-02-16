import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('UNCONV-004: @TestFactory', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'UNCONV-004', { minConfidence: 0 });
  });
});

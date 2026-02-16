import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('UNCONVERTIBLE-004: assert.throws', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'UNCONVERTIBLE-004', { minConfidence: 0 });
  });
});

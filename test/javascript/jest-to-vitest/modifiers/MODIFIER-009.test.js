import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MODIFIER-009: Sequential/concurrent execution', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MODIFIER-009');
  });
});

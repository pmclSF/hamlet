import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('STRUCT-004: multiple test methods in one class', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'STRUCT-004');
  });
});

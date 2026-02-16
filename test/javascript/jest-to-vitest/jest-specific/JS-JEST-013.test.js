import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('JS-JEST-013: Custom test environment (jsdom)', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'JS-JEST-013');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('FIXTURE-004: setUp with multi-line body', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'FIXTURE-004');
  });
});

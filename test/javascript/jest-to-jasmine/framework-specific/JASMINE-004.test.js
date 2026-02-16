import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('JASMINE-004: test.only to fit, test.skip to xit', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'JASMINE-004');
  });
});

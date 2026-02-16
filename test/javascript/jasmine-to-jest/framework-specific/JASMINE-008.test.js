import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('JASMINE-008: Combined clock and spy', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'JASMINE-008');
  });
});

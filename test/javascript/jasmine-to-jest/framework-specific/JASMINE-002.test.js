import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('JASMINE-002: jasmine.clock().mockDate', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'JASMINE-002');
  });
});

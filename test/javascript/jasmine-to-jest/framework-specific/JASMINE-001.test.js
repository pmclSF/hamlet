import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('JASMINE-001: jasmine.clock install/tick/uninstall', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'JASMINE-001');
  });
});

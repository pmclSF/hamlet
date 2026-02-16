import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('IMPORT-003: Mixed JUnit and non-JUnit imports', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'IMPORT-003');
  });
});

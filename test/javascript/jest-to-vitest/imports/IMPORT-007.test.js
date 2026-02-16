import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('IMPORT-007: Global setup file reference (jest.setup.js patterns)', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'IMPORT-007');
  });
});

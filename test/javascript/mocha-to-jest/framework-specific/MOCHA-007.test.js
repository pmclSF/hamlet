import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCHA-007: to.exist to toBeDefined', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCHA-007');
  });
});

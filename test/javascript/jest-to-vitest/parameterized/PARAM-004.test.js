import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('PARAM-004: Table-driven tests with tagged template literal', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'PARAM-004');
  });
});

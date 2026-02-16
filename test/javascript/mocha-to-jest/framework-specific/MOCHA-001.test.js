import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCHA-001: this.timeout to HAMLET-WARNING', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCHA-001');
  });
});

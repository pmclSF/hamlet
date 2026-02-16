import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('OUTPUT-001: Assert on stdout via console.log spy', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'OUTPUT-001');
  });
});

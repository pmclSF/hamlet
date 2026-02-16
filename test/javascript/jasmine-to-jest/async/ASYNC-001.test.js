import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ASYNC-001: done callback pass-through', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ASYNC-001');
  });
});

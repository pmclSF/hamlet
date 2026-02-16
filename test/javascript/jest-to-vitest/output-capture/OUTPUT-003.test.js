import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('OUTPUT-003: Suppress output during tests', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'OUTPUT-003');
  });
});

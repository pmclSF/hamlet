import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('DECOR-003: multiple nose decorators and assertions', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'DECOR-003');
  });
});

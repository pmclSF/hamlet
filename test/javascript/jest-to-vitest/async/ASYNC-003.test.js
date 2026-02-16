import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ASYNC-003: Callback-based test with done', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ASYNC-003');
  });
});

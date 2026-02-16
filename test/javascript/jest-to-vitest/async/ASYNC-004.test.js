import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ASYNC-004: Timer mocking with fake timers', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ASYNC-004');
  });
});

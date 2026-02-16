import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCK-010: .calls.mostRecent().args to .mock.lastCall', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCK-010');
  });
});

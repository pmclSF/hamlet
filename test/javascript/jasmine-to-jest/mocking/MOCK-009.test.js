import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCK-009: .calls.argsFor(n) to .mock.calls[n]', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCK-009');
  });
});

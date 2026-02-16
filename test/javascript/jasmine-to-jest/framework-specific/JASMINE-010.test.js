import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('JASMINE-010: .calls.first().args to .mock.calls[0]', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'JASMINE-010');
  });
});

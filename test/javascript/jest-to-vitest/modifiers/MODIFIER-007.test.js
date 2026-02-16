import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MODIFIER-007: Retry on failure (jest.retryTimes)', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MODIFIER-007');
  });
});

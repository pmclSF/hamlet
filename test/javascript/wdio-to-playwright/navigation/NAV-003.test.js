import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('NAV-003: pause to waitForTimeout', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'NAV-003');
  });
});

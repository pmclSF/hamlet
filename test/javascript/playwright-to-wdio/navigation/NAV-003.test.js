import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('NAV-003: waitForTimeout to pause', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'NAV-003');
  });
});

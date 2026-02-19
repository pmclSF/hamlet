import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ACT-003: locator.waitFor to waitForSelector', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ACT-003');
  });
});

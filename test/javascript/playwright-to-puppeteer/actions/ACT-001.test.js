import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ACT-001: locator.fill to page.type', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ACT-001');
  });
});

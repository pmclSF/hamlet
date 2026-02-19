import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ACT-001: page.type to locator.fill', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ACT-001');
  });
});

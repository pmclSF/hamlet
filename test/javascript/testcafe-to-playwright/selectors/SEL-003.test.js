import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('SEL-003: Selector properties (exists, visible, count)', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'SEL-003');
  });
});

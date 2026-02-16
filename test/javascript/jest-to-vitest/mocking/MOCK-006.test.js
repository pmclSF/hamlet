import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCK-006: Partial module mock with requireActual', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCK-006');
  });
});

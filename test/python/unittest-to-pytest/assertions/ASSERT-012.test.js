import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ASSERT-012: assertNotIsInstance to assert not isinstance', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ASSERT-012');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ASSERT-013: assertCountEqual to assert sorted == sorted', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ASSERT-013');
  });
});

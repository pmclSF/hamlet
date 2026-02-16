import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ASSERT-017: Chained/multiple assertions on same value', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ASSERT-017');
  });
});

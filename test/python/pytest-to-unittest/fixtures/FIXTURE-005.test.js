import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('FIXTURE-005: Fixture with params to HAMLET-TODO', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'FIXTURE-005');
  });
});

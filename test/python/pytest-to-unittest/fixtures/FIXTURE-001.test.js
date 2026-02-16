import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('FIXTURE-001: autouse fixture to setUp', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'FIXTURE-001');
  });
});

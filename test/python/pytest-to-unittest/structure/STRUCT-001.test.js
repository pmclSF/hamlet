import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('STRUCT-001: Single bare test function to class-wrapped method', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'STRUCT-001');
  });
});

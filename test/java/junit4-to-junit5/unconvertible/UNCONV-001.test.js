import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('UNCONV-001: ExpectedException @Rule', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'UNCONV-001', { minConfidence: 0 });
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('PARAM-002: Complex constructor injection', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'PARAM-002', { minConfidence: 0 });
  });
});

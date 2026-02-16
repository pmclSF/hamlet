import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('CONTEXT-006: Factory functions for test data', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'CONTEXT-006');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('CONTEXT-007: beforeEach setting instance variables', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'CONTEXT-007');
  });
});

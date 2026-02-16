import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('PARAM-002: it.each with multiple arguments', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'PARAM-002');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ERROR-004: Try/catch inside test body', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ERROR-004');
  });
});

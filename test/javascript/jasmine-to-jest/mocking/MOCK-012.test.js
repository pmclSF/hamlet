import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCK-012: Combined spyOn with returnValue and calledWith', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCK-012');
  });
});

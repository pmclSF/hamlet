import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCK-009: not.toHaveBeenCalled to to.not.have.been.called', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCK-009');
  });
});

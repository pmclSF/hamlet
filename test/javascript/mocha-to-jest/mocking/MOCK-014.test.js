import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCK-014: chai-sinon expect(fn).to.have.been.calledWith', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCK-014');
  });
});

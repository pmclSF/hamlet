import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCK-008: sinon.assert.calledWith to toHaveBeenCalledWith', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCK-008');
  });
});

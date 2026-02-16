import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCK-007: sinon.assert.calledOnce to toHaveBeenCalledTimes(1)', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCK-007');
  });
});

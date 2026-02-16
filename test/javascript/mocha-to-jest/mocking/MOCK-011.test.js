import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCK-011: sinon.restore to jest.restoreAllMocks', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCK-011');
  });
});

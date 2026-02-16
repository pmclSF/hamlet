import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCHA-008: to.be.an.instanceOf to toBeInstanceOf', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCHA-008');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('ASSERT-009: to.be.a(type) to typeof check', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'ASSERT-009');
  });
});

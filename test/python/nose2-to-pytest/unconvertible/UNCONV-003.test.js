import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('UNCONV-003: such DSL marked as HAMLET-TODO', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'UNCONV-003');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('FIXTURE-004: Session-scoped fixture to HAMLET-TODO', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'FIXTURE-004');
  });
});

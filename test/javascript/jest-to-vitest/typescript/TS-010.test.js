import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('TS-010: Declaration merging / module augmentation in test', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'TS-010', { ext: '.ts' });
  });
});

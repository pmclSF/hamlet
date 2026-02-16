import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('OUTPUT-005: Capture and assert on multiple log levels', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'OUTPUT-005');
  });
});

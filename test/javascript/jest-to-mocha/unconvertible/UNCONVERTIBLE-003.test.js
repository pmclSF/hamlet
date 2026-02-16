import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('UNCONVERTIBLE-003: jest.mock with jest.fn', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'UNCONVERTIBLE-003', { minConfidence: 0 });
  });
});

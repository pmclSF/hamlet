import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('UNCONVERTIBLE-002: jest.mock with virtual module', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'UNCONVERTIBLE-002', { minConfidence: 50 });
  });
});

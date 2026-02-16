import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('UNCONVERTIBLE-005: jest.mock with mocking and assertions', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'UNCONVERTIBLE-005', { minConfidence: 0 });
  });
});

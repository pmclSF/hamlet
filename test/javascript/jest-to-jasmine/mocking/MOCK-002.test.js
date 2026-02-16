import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCK-002: jest.spyOn to spyOn', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCK-002');
  });
});

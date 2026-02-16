import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCK-001: jest.fn to jasmine.createSpy', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCK-001');
  });
});

import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCK-002: jasmine.createSpyObj to object with jest.fn', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'MOCK-002');
  });
});

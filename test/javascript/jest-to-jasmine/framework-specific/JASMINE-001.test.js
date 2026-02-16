import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('JASMINE-001: jest.useFakeTimers to jasmine.clock().install()', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'JASMINE-001');
  });
});

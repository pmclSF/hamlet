import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('PARAM-003: DataProvider with multiple consumers', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'PARAM-003');
  });
});

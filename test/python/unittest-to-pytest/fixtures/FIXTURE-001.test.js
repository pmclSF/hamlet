import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('FIXTURE-001: setUp converted to pytest.fixture', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'FIXTURE-001');
  });
});

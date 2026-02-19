import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('NAV-001: .page to cy.visit in beforeEach', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'NAV-001');
  });
});

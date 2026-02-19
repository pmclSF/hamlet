import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('NAV-002: t.navigateTo to cy.visit', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'NAV-002');
  });
});

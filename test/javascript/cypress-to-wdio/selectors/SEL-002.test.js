import { fileURLToPath } from 'url';
import path from 'path';
import { runFixture } from '../convert.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('SEL-002: cy.contains to $(*=text)', () => {
  it('should convert correctly', async () => {
    await runFixture(__dirname, 'SEL-002');
  });
});

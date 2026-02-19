import { CypressToPlaywright } from '../../src/converters/CypressToPlaywright.js';

describe('CypressToPlaywright — residual patterns', () => {
  let converter;

  beforeEach(() => {
    converter = new CypressToPlaywright();
  });

  describe('convert', () => {
    describe('cy.getBySel()', () => {
      it('should convert cy.getBySel() to page.getByTestId()', async () => {
        const input = `cy.getBySel('submit-btn');`;
        const result = await converter.convert(input);
        expect(result).toContain("page.getByTestId('submit-btn')");
        expect(result).not.toContain('cy.getBySel');
      });

      it('should convert cy.getBySel() with a variable argument', async () => {
        const input = `cy.getBySel(selectorVar);`;
        const result = await converter.convert(input);
        expect(result).toContain('page.getByTestId(selectorVar)');
        expect(result).not.toContain('cy.getBySel');
      });
    });

    describe('cy.getBySelLike()', () => {
      it('should convert cy.getBySelLike() to page.locator with data-test*= selector', async () => {
        const input = `cy.getBySelLike('user-item');`;
        const result = await converter.convert(input);
        expect(result).toContain('page.locator(`[data-test*=');
        expect(result).not.toContain('cy.getBySelLike');
      });
    });

    describe('cy.location()', () => {
      it('should convert cy.location() to new URL(page.url())', async () => {
        const input = `cy.location();`;
        const result = await converter.convert(input);
        expect(result).toContain('new URL(page.url())');
        expect(result).not.toContain('cy.location()');
      });

      it('should convert cy.location("pathname") to new URL(page.url()).pathname', async () => {
        const input = `cy.location('pathname');`;
        const result = await converter.convert(input);
        expect(result).toContain('new URL(page.url()).pathname');
        expect(result).not.toContain('cy.location');
      });

      it('should convert cy.location("pathname").should("eq", "/path") to toHaveURL', async () => {
        const input = `cy.location('pathname').should('eq', '/dashboard');`;
        const result = await converter.convert(input);
        expect(result).toContain('await expect(page).toHaveURL');
        expect(result).toContain('/dashboard');
        expect(result).not.toContain('cy.location');
      });
    });

    describe('cy.getCookie()', () => {
      it('should convert cy.getCookie() to context.cookies() lookup', async () => {
        const input = `cy.getCookie('session_id');`;
        const result = await converter.convert(input);
        expect(result).toContain('await context.cookies()');
        expect(result).toContain('session_id');
        expect(result).not.toContain('cy.getCookie');
      });
    });

    describe('cy.getCookies()', () => {
      it('should convert cy.getCookies() to await context.cookies()', async () => {
        const input = `cy.getCookies();`;
        const result = await converter.convert(input);
        expect(result).toContain('await context.cookies()');
        expect(result).not.toContain('cy.getCookies');
      });
    });

    describe('cy.visualSnapshot()', () => {
      it('should convert cy.visualSnapshot() to page.screenshot()', async () => {
        const input = `cy.visualSnapshot('homepage');`;
        const result = await converter.convert(input);
        expect(result).toContain('await page.screenshot(');
        expect(result).not.toContain('cy.visualSnapshot');
      });
    });

    describe('cy.intercept — two-arg with .as()', () => {
      it('should convert cy.intercept(method, url).as(alias) to page.route with continue', async () => {
        const input = `cy.intercept('GET', '/api/users').as('getUsers');`;
        const result = await converter.convert(input);
        expect(result).toContain('await page.route');
        expect(result).toContain('/api/users');
        expect(result).toContain('route => route.continue()');
        expect(result).not.toContain('cy.intercept');
      });
    });

    describe('cy.intercept — three-arg with .as()', () => {
      it('should convert cy.intercept(method, url, response).as(alias) to page.route with fulfill', async () => {
        const input = `cy.intercept('POST', '/api/data', { body: 'ok' }).as('stub');`;
        const result = await converter.convert(input);
        expect(result).toContain('await page.route');
        expect(result).toContain('/api/data');
        expect(result).toContain('route.fulfill');
        expect(result).not.toContain('cy.intercept');
      });
    });

    describe('cy.task() — custom command (no Playwright equivalent)', () => {
      it('should leave cy.task() in the output since there is no automatic conversion', async () => {
        const input = `cy.task('seed:db');`;
        const result = await converter.convert(input);
        // cy.task has no Playwright equivalent — the converter does not
        // have a matching pattern, so it passes through unchanged
        expect(result).toContain('cy.task');
      });
    });

    describe('cy.database() — custom command (no Playwright equivalent)', () => {
      it('should leave cy.database() in the output since there is no automatic conversion', async () => {
        const input = `cy.database('seed');`;
        const result = await converter.convert(input);
        // cy.database has no Playwright equivalent — the converter does
        // not have a matching pattern, so it passes through unchanged
        expect(result).toContain('cy.database');
      });
    });
  });
});

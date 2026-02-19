describe('url assertions', () => {
  it('should check url', () => {
    cy.visit('/dashboard');
    cy.url().should('eq', 'http://localhost/dashboard');
    cy.url().should('include', '/dash');
  });
});

describe('Home Page', () => {
  it('should display the hero banner', () => {
    cy.visit('/');
    cy.get('[data-testid="hero-banner"]').should('be.visible');
  });

  it('should navigate to the products page', () => {
    cy.visit('/');
    cy.get('nav a[href="/products"]').click();
    cy.url().should('include', '/products');
    cy.get('h1').should('have.text', 'Our Products');
  });
});

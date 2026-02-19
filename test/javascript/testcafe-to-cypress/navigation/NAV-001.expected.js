describe('Navigation', () => {
  beforeEach(() => {
    cy.visit('http://localhost/home');
  });

it('should load page', () => {
  cy.get('#content').should('be.visible');
});

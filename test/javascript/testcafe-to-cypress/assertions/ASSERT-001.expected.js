describe('Assertions', () => {
  beforeEach(() => {
    cy.visit('http://localhost');
  });

it('should check visibility', () => {
  cy.get('#visible').should('be.visible');
  cy.get('#present').should('exist');
});

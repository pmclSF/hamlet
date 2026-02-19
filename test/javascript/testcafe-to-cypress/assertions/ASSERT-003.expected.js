describe('Count Assertions', () => {
  beforeEach(() => {
    cy.visit('http://localhost');
  });

it('should check count', () => {
  cy.get('.items').should('have.length', 5);
});

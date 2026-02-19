describe('Text Assertions', () => {
  beforeEach(() => {
    cy.visit('http://localhost');
  });

it('should check text', () => {
  cy.get('#msg').should('have.text', 'Hello');
  cy.get('#msg').should('contain', 'Hel');
});

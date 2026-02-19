describe('Value Assertions', () => {
  beforeEach(() => {
    cy.visit('http://localhost/form');
  });

it('should check value', () => {
  cy.get('#input').type('test');
  cy.get('#input').should('have.value', 'test');
});

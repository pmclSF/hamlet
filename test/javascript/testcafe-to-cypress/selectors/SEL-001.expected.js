describe('Selectors', () => {
  beforeEach(() => {
    cy.visit('http://localhost/form');
  });

it('should find elements', () => {
  cy.get('#name').type('John');
  cy.get('#submit').click();
});

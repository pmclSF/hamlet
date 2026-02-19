describe('My Suite', () => {
  beforeEach(() => {
    cy.visit('http://localhost');
  });

it('first test', () => {
  cy.get('#btn').click();
});

it('second test', () => {
  cy.get('#other').click();
});

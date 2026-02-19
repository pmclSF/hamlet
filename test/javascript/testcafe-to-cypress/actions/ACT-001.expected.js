describe('Actions', () => {
  beforeEach(() => {
    cy.visit('http://localhost/form');
  });

it('should type text', () => {
  cy.get('#email').type('user@test.com');
  cy.get('#password').type('secret');
});

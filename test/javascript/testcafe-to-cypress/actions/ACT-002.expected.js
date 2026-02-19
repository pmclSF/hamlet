describe('Click Actions', () => {
  beforeEach(() => {
    cy.visit('http://localhost/app');
  });

it('should click', () => {
  cy.get('#submit').click();
  cy.get('#double').dblclick();
});

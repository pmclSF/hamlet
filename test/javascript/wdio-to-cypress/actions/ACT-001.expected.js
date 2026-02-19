describe('form actions', () => {
  it('should type values', () => {
    cy.visit('/form');
    cy.get('#email').clear().type('user@test.com');
    cy.get('#name').clear();
  });
});

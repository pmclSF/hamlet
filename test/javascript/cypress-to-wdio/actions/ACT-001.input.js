describe('form actions', () => {
  it('should type values', () => {
    cy.visit('/form');
    cy.get('#email').type('user@test.com');
    cy.get('#field').clear();
  });
});

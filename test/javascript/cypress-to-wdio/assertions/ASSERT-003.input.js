describe('text assertions', () => {
  it('should check text', () => {
    cy.visit('/page');
    cy.get('#msg').should('have.text', 'Hello');
    cy.get('#msg').should('contain', 'World');
    cy.get('#input').should('have.value', 'test');
  });
});

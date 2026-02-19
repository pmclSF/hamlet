describe('visibility', () => {
  it('should check visibility', () => {
    cy.visit('/page');
    cy.get('#elem').should('be.visible');
    cy.get('#hidden').should('not.be.visible');
  });
});

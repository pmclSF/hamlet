describe('click actions', () => {
  it('should click', () => {
    cy.visit('/app');
    cy.get('#btn').click();
    cy.get('#dbl').dblclick();
  });
});

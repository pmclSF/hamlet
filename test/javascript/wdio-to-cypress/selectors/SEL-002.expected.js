describe('text selectors', () => {
  it('should find by text', () => {
    cy.visit('/home');
    cy.contains('Submit').click();
  });
});

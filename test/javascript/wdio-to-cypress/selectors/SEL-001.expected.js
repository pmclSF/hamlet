describe('selectors', () => {
  it('should find elements', () => {
    cy.visit('/form');
    cy.get('#username').clear().type('test');
    cy.get('#submit').click();
  });
});

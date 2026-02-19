describe('selectors', () => {
  it('should find elements', () => {
    cy.visit('/form');
    cy.get('#username').type('test');
    cy.get('#submit').click();
  });
});

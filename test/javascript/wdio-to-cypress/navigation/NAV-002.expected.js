describe('navigation', () => {
  it('should navigate around', () => {
    cy.visit('/page1');
    cy.reload();
    cy.go('back');
    cy.go('forward');
  });
});

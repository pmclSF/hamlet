describe('async addition', () => {
  beforeEach(() => {
    cy.visit('/setup');
  });

  it('should add async', () => {
    cy.get('#btn').click();
  });
});

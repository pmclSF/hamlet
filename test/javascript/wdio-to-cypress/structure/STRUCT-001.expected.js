describe('async removal', () => {
  beforeEach(() => {
    cy.visit('/setup');
  });

  it('should remove async', () => {
    cy.get('#btn').click();
  });
});

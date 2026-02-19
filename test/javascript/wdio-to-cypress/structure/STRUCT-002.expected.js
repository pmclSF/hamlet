describe('Login Flow', () => {
  beforeEach(() => {
    cy.visit('/login');
  });

  it('should login', () => {
    cy.get('#username').clear().type('admin');
    cy.get('#password').clear().type('pass123');
    cy.get('#login-btn').click();
    cy.url().should('eq', 'http://localhost/dashboard');
    cy.get('#welcome').should('be.visible');
  });
});

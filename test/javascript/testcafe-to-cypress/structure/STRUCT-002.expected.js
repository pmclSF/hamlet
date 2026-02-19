describe('Login Flow', () => {
  beforeEach(() => {
    cy.visit('http://localhost/login');
  });

it('should login successfully', () => {
  cy.get('#username').type('admin');
  cy.get('#password').type('pass123');
  cy.get('#login-btn').click();
  cy.get('#welcome').should('be.visible');
  cy.get('#welcome').should('have.text', 'Welcome, admin');
});

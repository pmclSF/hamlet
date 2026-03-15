describe('Login Flow', () => {
  beforeEach(() => {
    cy.visit('/login');
  });

  it('should display login form', () => {
    cy.get('form').should('be.visible');
    cy.get('input[name="email"]').should('exist');
    cy.get('input[name="password"]').should('exist');
  });

  it('should login with valid credentials', () => {
    cy.get('input[name="email"]').type('user@example.com');
    cy.get('input[name="password"]').type('password123');
    cy.get('button[type="submit"]').click();
    cy.url().should('include', '/dashboard');
  });

  it('should show error for invalid credentials', () => {
    cy.get('input[name="email"]').type('bad@example.com');
    cy.get('input[name="password"]').type('wrong');
    cy.get('button[type="submit"]').click();
    cy.get('.error').should('be.visible');
  });
});

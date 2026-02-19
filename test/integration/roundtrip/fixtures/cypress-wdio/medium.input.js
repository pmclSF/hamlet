describe('User Registration', () => {
  beforeEach(() => {
    cy.visit('/register');
  });

  it('should display the registration form', () => {
    cy.get('[data-testid="register-form"]').should('be.visible');
    cy.get('[data-testid="email-input"]').should('exist');
    cy.get('[data-testid="password-input"]').should('exist');
  });

  it('should show validation errors for empty submission', () => {
    cy.get('[data-testid="submit-button"]').click();
    cy.get('[data-testid="email-error"]').should('contain.text', 'Email is required');
    cy.get('[data-testid="password-error"]').should('contain.text', 'Password is required');
  });

  it('should accept valid input and submit', () => {
    cy.get('[data-testid="email-input"]').type('newuser@example.com');
    cy.get('[data-testid="password-input"]').type('Str0ng!Pass');
    cy.get('[data-testid="submit-button"]').click();
    cy.url().should('include', '/dashboard');
  });

  it('should show an error for duplicate email', () => {
    cy.get('[data-testid="email-input"]').type('existing@example.com');
    cy.get('[data-testid="password-input"]').type('Str0ng!Pass');
    cy.get('[data-testid="submit-button"]').click();
    cy.get('[data-testid="form-error"]').should('contain.text', 'already registered');
  });
});

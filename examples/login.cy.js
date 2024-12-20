describe('Login Test', () => {
    it('should login successfully', () => {
      cy.visit('/login');
      cy.get('[data-test=username]').type('testuser');
      cy.get('[data-test=password]').type('password123');
      cy.get('[data-test=submit]').click();
      cy.get('.welcome-message').should('be.visible');
      cy.get('.welcome-message').should('have.text', 'Welcome, testuser!');
    });
  });
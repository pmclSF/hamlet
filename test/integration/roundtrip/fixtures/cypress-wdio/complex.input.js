// E2E tests for the admin dashboard management features
describe('Admin Dashboard', () => {
  beforeEach(() => {
    cy.intercept('GET', '/api/admin/stats', { fixture: 'admin-stats.json' }).as('getStats');
    cy.intercept('GET', '/api/admin/users*', { fixture: 'admin-users.json' }).as('getUsers');
    cy.visit('/admin');
    cy.wait('@getStats');
  });

  describe('overview panel', () => {
    it('should display key metrics', () => {
      cy.get('[data-testid="metric-revenue"]').should('contain.text', '$12,345');
      cy.get('[data-testid="metric-orders"]').should('contain.text', '89');
      cy.get('[data-testid="metric-users"]').should('contain.text', '1,204');
    });

    it('should render the revenue chart', () => {
      cy.get('[data-testid="revenue-chart"] canvas').should('be.visible');
    });
  });

  describe('user management', () => {
    beforeEach(() => {
      cy.get('[data-testid="nav-users"]').click();
      cy.wait('@getUsers');
    });

    it('should list users in a table', () => {
      cy.get('[data-testid="user-table"] tbody tr').should('have.length.greaterThan', 0);
    });

    it('should search users by name', () => {
      cy.intercept('GET', '/api/admin/users?q=jane*', { fixture: 'admin-users-filtered.json' }).as('searchUsers');
      cy.get('[data-testid="user-search"]').type('jane');
      cy.wait('@searchUsers');
      cy.get('[data-testid="user-table"] tbody tr').should('have.length', 1);
      cy.get('[data-testid="user-table"] tbody tr').first().should('contain.text', 'Jane');
    });

    it('should open the user detail modal', () => {
      cy.get('[data-testid="user-table"] tbody tr').first().find('[data-testid="view-button"]').click();
      cy.get('[data-testid="user-modal"]').should('be.visible');
      cy.get('[data-testid="user-modal"] .user-email').should('not.be.empty');
    });

    it('should deactivate a user account', () => {
      cy.intercept('PATCH', '/api/admin/users/*/status', { statusCode: 200 }).as('updateStatus');
      cy.get('[data-testid="user-table"] tbody tr').first().find('[data-testid="view-button"]').click();
      cy.get('[data-testid="deactivate-button"]').click();
      cy.get('[data-testid="confirm-dialog"] [data-testid="confirm-yes"]').click();
      cy.wait('@updateStatus');
      cy.get('[data-testid="user-modal"]').should('contain.text', 'Inactive');
    });

    it('should paginate through user results', () => {
      cy.intercept('GET', '/api/admin/users?page=2', { fixture: 'admin-users-page2.json' }).as('page2');
      cy.get('[data-testid="pagination-next"]').click();
      cy.wait('@page2');
      cy.get('[data-testid="pagination-current"]').should('have.text', '2');
    });
  });
});

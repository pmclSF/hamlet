describe('Product Search', () => {
  beforeEach(() => {
    cy.intercept('GET', '/api/products*', { fixture: 'products.json' }).as('getProducts');
    cy.visit('/products');
    cy.wait('@getProducts');
  });

  it('should display a list of products', () => {
    cy.get('[data-testid="product-card"]').should('have.length.greaterThan', 0);
  });

  it('should filter products by search term', () => {
    cy.get('[data-testid="search-input"]').type('wireless');
    cy.get('[data-testid="product-card"]').each(($card) => {
      cy.wrap($card).should('contain.text', 'wireless');
    });
  });

  it('should show product details on click', () => {
    cy.get('[data-testid="product-card"]').first().click();
    cy.get('[data-testid="product-detail"]').should('be.visible');
    cy.get('[data-testid="product-price"]').should('not.be.empty');
  });

  it('should add a product to the cart', () => {
    cy.get('[data-testid="product-card"]').first().find('button.add-to-cart').click();
    cy.get('[data-testid="cart-badge"]').should('contain.text', '1');
  });
});

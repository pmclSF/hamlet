// Cypress E2E test for an e-commerce checkout flow
// Inspired by real-world Cypress tests for online stores

describe('Checkout Flow', () => {
  beforeEach(() => {
    cy.intercept('GET', '/api/products*', { fixture: 'products.json' }).as('getProducts');
    cy.intercept('GET', '/api/cart', { fixture: 'cart.json' }).as('getCart');
    cy.intercept('POST', '/api/cart/items', { statusCode: 201, body: { success: true } }).as('addToCart');
    cy.intercept('POST', '/api/orders', {
      statusCode: 201,
      body: { orderId: 'ORD-12345', status: 'confirmed' },
    }).as('placeOrder');
    cy.intercept('GET', '/api/shipping/rates', {
      body: [
        { id: 'standard', label: 'Standard (5-7 days)', price: 5.99 },
        { id: 'express', label: 'Express (1-2 days)', price: 14.99 },
      ],
    }).as('getShippingRates');

    cy.visit('/shop');
    cy.wait('@getProducts');
  });

  describe('Product browsing', () => {
    it('should display product cards with prices', () => {
      cy.get('[data-testid="product-card"]').should('have.length.greaterThan', 0);
      cy.get('[data-testid="product-card"]').first().within(() => {
        cy.get('.product-name').should('be.visible');
        cy.get('.product-price').should('contain', '$');
      });
    });

    it('should filter products by category', () => {
      cy.get('[data-testid="category-filter"]').select('Electronics');
      cy.get('[data-testid="product-card"]').each(($card) => {
        cy.wrap($card).find('.product-category').should('have.text', 'Electronics');
      });
    });

    it('should open product detail page on card click', () => {
      cy.get('[data-testid="product-card"]').first().click();
      cy.url().should('include', '/product/');
      cy.get('[data-testid="product-detail"]').should('be.visible');
      cy.get('[data-testid="add-to-cart-btn"]').should('be.visible');
    });
  });

  describe('Shopping cart', () => {
    it('should add a product to the cart', () => {
      cy.get('[data-testid="product-card"]').first().click();
      cy.get('[data-testid="quantity-input"]').clear().type('2');
      cy.get('[data-testid="add-to-cart-btn"]').click();

      cy.wait('@addToCart').its('request.body').should('deep.include', { quantity: 2 });
      cy.get('[data-testid="cart-badge"]').should('contain', '2');
    });

    it('should show the cart summary in the sidebar', () => {
      cy.get('[data-testid="cart-icon"]').click();
      cy.wait('@getCart');
      cy.get('[data-testid="cart-sidebar"]').should('be.visible');
      cy.get('[data-testid="cart-item"]').should('have.length.greaterThan', 0);
      cy.get('[data-testid="cart-total"]').should('not.be.empty');
    });

    it('should remove an item from the cart', () => {
      cy.intercept('DELETE', '/api/cart/items/*', { statusCode: 200 }).as('removeItem');
      cy.get('[data-testid="cart-icon"]').click();
      cy.wait('@getCart');
      cy.get('[data-testid="remove-item-btn"]').first().click();
      cy.wait('@removeItem');
      cy.get('[data-testid="cart-item"]').should('have.length', 0);
    });
  });

  describe('Checkout process', () => {
    beforeEach(() => {
      cy.visit('/checkout');
      cy.wait('@getCart');
      cy.wait('@getShippingRates');
    });

    it('should display shipping address form fields', () => {
      cy.get('[data-testid="shipping-form"]').within(() => {
        cy.get('input[name="firstName"]').should('be.visible');
        cy.get('input[name="lastName"]').should('be.visible');
        cy.get('input[name="address"]').should('be.visible');
        cy.get('input[name="city"]').should('be.visible');
        cy.get('select[name="state"]').should('be.visible');
        cy.get('input[name="zip"]').should('be.visible');
      });
    });

    it('should validate required fields before proceeding', () => {
      cy.get('[data-testid="continue-btn"]').click();
      cy.get('.field-error').should('have.length.greaterThan', 0);
      cy.contains('First name is required').should('be.visible');
    });

    it('should allow selecting a shipping method', () => {
      cy.get('[data-testid="shipping-option"]').should('have.length', 2);
      cy.get('[data-testid="shipping-option"]').last().click();
      cy.get('[data-testid="shipping-cost"]').should('contain', '$14.99');
    });

    it('should place an order after filling out all fields', () => {
      cy.get('input[name="firstName"]').type('Jane');
      cy.get('input[name="lastName"]').type('Smith');
      cy.get('input[name="address"]').type('123 Main St');
      cy.get('input[name="city"]').type('Portland');
      cy.get('select[name="state"]').select('OR');
      cy.get('input[name="zip"]').type('97201');
      cy.get('[data-testid="shipping-option"]').first().click();
      cy.get('[data-testid="place-order-btn"]').click();

      cy.wait('@placeOrder');
      cy.url().should('include', '/order/confirmation');
      cy.contains('ORD-12345').should('be.visible');
      cy.contains('Order confirmed').should('be.visible');
    });
  });
});

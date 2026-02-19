// E2E tests for the checkout flow with authenticated user
describe('Checkout Flow', () => {
  beforeEach(() => {
    cy.session('authenticated-user', () => {
      cy.visit('/login');
      cy.get('[data-testid="email-input"]').type('buyer@example.com');
      cy.get('[data-testid="password-input"]').type('securePass123');
      cy.get('[data-testid="login-button"]').click();
      cy.url().should('not.include', '/login');
    });
  });

  describe('cart management', () => {
    beforeEach(() => {
      cy.intercept('GET', '/api/cart', { fixture: 'cart-with-items.json' }).as('getCart');
      cy.intercept('GET', '/api/products/*', { fixture: 'product-detail.json' }).as('getProduct');
      cy.visit('/cart');
      cy.wait('@getCart');
    });

    it('should display cart items with correct quantities', () => {
      cy.get('[data-testid="cart-item"]').should('have.length', 3);
      cy.get('[data-testid="cart-item"]').first().find('.quantity-display').should('have.text', '2');
    });

    it('should update the quantity of an item', () => {
      cy.intercept('PATCH', '/api/cart/items/*', { statusCode: 200 }).as('updateItem');
      cy.get('[data-testid="cart-item"]').first().find('.quantity-increase').click();
      cy.wait('@updateItem');
      cy.get('[data-testid="cart-item"]').first().find('.quantity-display').should('have.text', '3');
    });

    it('should remove an item from the cart', () => {
      cy.intercept('DELETE', '/api/cart/items/*', { statusCode: 204 }).as('deleteItem');
      cy.get('[data-testid="cart-item"]').first().find('[data-testid="remove-button"]').click();
      cy.wait('@deleteItem');
      cy.get('[data-testid="cart-item"]').should('have.length', 2);
    });

    it('should display the correct subtotal', () => {
      cy.get('[data-testid="cart-subtotal"]').should('contain.text', '$149.97');
    });
  });

  describe('payment and confirmation', () => {
    beforeEach(() => {
      cy.intercept('GET', '/api/cart', { fixture: 'cart-with-items.json' }).as('getCart');
      cy.intercept('POST', '/api/orders', {
        statusCode: 201,
        body: { orderId: 'ord-789', status: 'confirmed' },
      }).as('placeOrder');
      cy.visit('/checkout');
      cy.wait('@getCart');
    });

    it('should fill in shipping details', () => {
      cy.get('[data-testid="shipping-name"]').type('Jane Doe');
      cy.get('[data-testid="shipping-address"]').type('123 Main St');
      cy.get('[data-testid="shipping-city"]').type('Springfield');
      cy.get('[data-testid="shipping-zip"]').type('62704');
      cy.get('[data-testid="continue-to-payment"]').click();
      cy.get('[data-testid="payment-section"]').should('be.visible');
    });

    it('should submit the order and show confirmation', () => {
      cy.fixture('shipping-details.json').then((shipping) => {
        cy.get('[data-testid="shipping-name"]').type(shipping.name);
        cy.get('[data-testid="shipping-address"]').type(shipping.address);
        cy.get('[data-testid="shipping-city"]').type(shipping.city);
        cy.get('[data-testid="shipping-zip"]').type(shipping.zip);
      });
      cy.get('[data-testid="continue-to-payment"]').click();
      cy.get('[data-testid="place-order-button"]').click();
      cy.wait('@placeOrder');
      cy.url().should('include', '/confirmation');
      cy.get('[data-testid="order-id"]').should('contain.text', 'ord-789');
    });

    it('should display an error when order submission fails', () => {
      cy.intercept('POST', '/api/orders', { statusCode: 500 }).as('failedOrder');
      cy.get('[data-testid="continue-to-payment"]').click();
      cy.get('[data-testid="place-order-button"]').click();
      cy.wait('@failedOrder');
      cy.get('[data-testid="error-message"]').should('be.visible');
      cy.get('[data-testid="error-message"]').should('contain.text', 'unable to process');
    });

    it('should allow the user to return to the cart from checkout', () => {
      cy.get('[data-testid="back-to-cart"]').click();
      cy.url().should('include', '/cart');
    });
  });
});

// E2E tests for the checkout flow
describe('Checkout Flow', () => {
  beforeEach(() => {
    cy.visit('/login');
    cy.get('[data-testid="email-input"]').type('buyer@example.com');
    cy.get('[data-testid="password-input"]').type('securePass123');
    cy.get('[data-testid="login-button"]').click();
    cy.url().should('not.include', '/login');
  });

  describe('cart management', () => {
    beforeEach(() => {
      cy.visit('/cart');
    });

    it('should display cart items with correct quantities', () => {
      cy.get('[data-testid="cart-item"]').should('have.length', 3);
      cy.get('[data-testid="cart-item"]').first().find('.quantity-display').should('have.text', '2');
    });

    it('should update the quantity of an item', () => {
      cy.get('[data-testid="cart-item"]').first().find('.quantity-increase').click();
      cy.get('[data-testid="cart-item"]').first().find('.quantity-display').should('have.text', '3');
    });

    it('should remove an item from the cart', () => {
      cy.get('[data-testid="cart-item"]').first().find('[data-testid="remove-button"]').click();
      cy.get('[data-testid="cart-item"]').should('have.length', 2);
    });

    it('should display the correct subtotal', () => {
      cy.get('[data-testid="cart-subtotal"]').should('contain.text', '$149.97');
    });
  });

  describe('payment and confirmation', () => {
    beforeEach(() => {
      cy.visit('/checkout');
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
      cy.get('[data-testid="shipping-name"]').type('Jane Doe');
      cy.get('[data-testid="shipping-address"]').type('123 Main St');
      cy.get('[data-testid="shipping-city"]').type('Springfield');
      cy.get('[data-testid="shipping-zip"]').type('62704');
      cy.get('[data-testid="continue-to-payment"]').click();
      cy.get('[data-testid="place-order-button"]').click();
      cy.url().should('include', '/confirmation');
      cy.get('[data-testid="order-id"]').should('be.visible');
    });

    it('should allow the user to return to the cart from checkout', () => {
      cy.get('[data-testid="back-to-cart"]').click();
      cy.url().should('include', '/cart');
    });
  });
});

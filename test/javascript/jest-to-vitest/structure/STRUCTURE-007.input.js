describe('PaymentProcessor', () => {
  it('should process credit card payments', () => {
    // TODO: implement once payment gateway SDK is integrated
    // Steps:
    // 1. Create a payment intent
    // 2. Attach payment method
    // 3. Confirm the payment
    // 4. Assert on the payment status
  });

  it('should handle refunds', () => {
    // Blocked by JIRA-4521: refund API not yet available in sandbox
    // Will need to mock the gateway response for now
  });
});

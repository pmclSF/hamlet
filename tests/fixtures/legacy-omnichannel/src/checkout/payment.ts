export function validatePayment(token: string) {
  return token.startsWith('pay_');
}

export function processPayment(amount: number, currency: string) {
  return { transactionId: 'txn_' + Date.now(), amount, currency, status: 'captured' };
}

export function refundPayment(transactionId: string, amount: number) {
  return { transactionId, refundAmount: amount, status: 'refunded' };
}

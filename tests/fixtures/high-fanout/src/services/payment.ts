export function processPayment(amount: number) {
  return { success: amount > 0, amount };
}

import { validatePayment } from './payment';
import { getUser, getCart } from '../shared/db-helper';

export function initiateCheckout(userId: string, cartId: string) {
  const user = getUser(userId);
  const cart = getCart(cartId);
  return { checkoutId: 'co_' + Date.now(), userId, cartId, status: 'initiated' };
}

export function completeCheckout(checkoutId: string, paymentToken: string) {
  const valid = validatePayment(paymentToken);
  return { checkoutId, status: valid ? 'completed' : 'failed' };
}

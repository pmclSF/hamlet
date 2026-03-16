const { getUser } = require('../shared/db-helper');

function addToCart(userId, productId, quantity) {
  const user = getUser(userId);
  return { cartId: 'cart_' + Date.now(), userId, items: [{ productId, quantity }] };
}

function removeFromCart(cartId, productId) {
  return { cartId, removed: productId };
}

function getCart(cartId) {
  return { cartId, items: [], total: 0 };
}

module.exports = { addToCart, removeFromCart, getCart };

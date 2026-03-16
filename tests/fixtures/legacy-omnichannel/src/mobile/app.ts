export function initMobileSession(deviceId: string, platform: string) {
  return { sessionId: 'mob_' + Date.now(), deviceId, platform };
}

export function submitMobileOrder(sessionId: string, cartId: string) {
  return { orderId: 'ord_' + Date.now(), sessionId, cartId, status: 'submitted' };
}

export function trackDelivery(orderId: string) {
  return { orderId, status: 'in_transit', eta: '2 days' };
}

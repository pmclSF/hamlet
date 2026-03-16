export function createRefund(orderId: string, amount: number) {
  return { refundId: 'ref_' + Date.now(), orderId, amount, status: 'pending' };
}

export function approveRefund(refundId: string) {
  return { refundId, status: 'approved' };
}

export function denyRefund(refundId: string, reason: string) {
  return { refundId, status: 'denied', reason };
}

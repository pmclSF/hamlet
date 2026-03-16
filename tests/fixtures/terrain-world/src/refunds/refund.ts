export function createRefund(chargeId: string, amount: number) {
  if (amount <= 0) throw new Error('Invalid refund amount');
  return { refundId: 'ref_' + Date.now(), chargeId, amount, status: 'pending' };
}

export function approveRefund(refundId: string) {
  return { refundId, status: 'approved' };
}

export function denyRefund(refundId: string, reason: string) {
  return { refundId, status: 'denied', reason };
}

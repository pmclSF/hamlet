export function createCharge(amount: number, currency: string) {
  if (amount <= 0) throw new Error('Invalid amount');
  return { chargeId: 'ch_' + Date.now(), amount, currency, status: 'pending' };
}

export function captureCharge(chargeId: string) {
  return { chargeId, status: 'captured' };
}

export function voidCharge(chargeId: string) {
  return { chargeId, status: 'voided' };
}

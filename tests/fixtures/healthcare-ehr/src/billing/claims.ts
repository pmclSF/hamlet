export function submitClaim(patientId: string, amount: number, procedure: string) {
  return { claimId: 'clm_' + Date.now(), patientId, amount, procedure, status: 'submitted' };
}
export function processClaim(claimId: string) { return { claimId, status: 'processed' }; }
export function denyClaim(claimId: string, reason: string) { return { claimId, status: 'denied', reason }; }

export function createInvoice(orgId: string, amount: number) {
  if (amount <= 0) throw new Error('Invalid amount');
  return { invoiceId: 'inv_' + Date.now(), orgId, amount, status: 'draft' };
}

export function finalizeInvoice(invoiceId: string) {
  return { invoiceId, status: 'finalized' };
}

export function voidInvoice(invoiceId: string) {
  return { invoiceId, status: 'voided' };
}

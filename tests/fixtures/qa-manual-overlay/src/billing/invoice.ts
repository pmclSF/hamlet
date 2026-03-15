export interface Invoice {
  id: string;
  customerId: string;
  items: { description: string; amount: number }[];
  status: 'draft' | 'sent' | 'paid' | 'overdue';
}

export function createInvoice(customerId: string, items: { description: string; amount: number }[]): Invoice {
  return { id: `inv_${Date.now()}`, customerId, items, status: 'draft' };
}

export function invoiceTotal(invoice: Invoice): number {
  return invoice.items.reduce((sum, item) => sum + item.amount, 0);
}

export function markAsPaid(invoice: Invoice): Invoice {
  return { ...invoice, status: 'paid' };
}

export function isOverdue(invoice: Invoice, daysSinceSent: number): boolean {
  return invoice.status === 'sent' && daysSinceSent > 30;
}

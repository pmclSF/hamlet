export function transactionsCreate(input: string) { return { id: 'transactions_' + Date.now(), input, status: 'created' }; }
export function transactionsGet(id: string) { return { id, data: {}, found: true }; }
export function transactionsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function transactionsDelete(id: string) { return { id, deleted: true }; }

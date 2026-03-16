export function paymentsCreate(input: string) { return { id: 'payments_' + Date.now(), input, status: 'created' }; }
export function paymentsGet(id: string) { return { id, data: {}, found: true }; }
export function paymentsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function paymentsDelete(id: string) { return { id, deleted: true }; }

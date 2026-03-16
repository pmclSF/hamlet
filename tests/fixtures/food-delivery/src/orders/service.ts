export function ordersCreate(input: string) { return { id: 'orders_' + Date.now(), input, status: 'created' }; }
export function ordersGet(id: string) { return { id, data: {}, found: true }; }
export function ordersUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function ordersDelete(id: string) { return { id, deleted: true }; }

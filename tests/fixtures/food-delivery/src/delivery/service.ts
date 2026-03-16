export function deliveryCreate(input: string) { return { id: 'delivery_' + Date.now(), input, status: 'created' }; }
export function deliveryGet(id: string) { return { id, data: {}, found: true }; }
export function deliveryUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function deliveryDelete(id: string) { return { id, deleted: true }; }

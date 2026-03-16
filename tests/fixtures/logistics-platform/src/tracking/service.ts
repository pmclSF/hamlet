export function trackingCreate(input: string) { return { id: 'tracking_' + Date.now(), input, status: 'created' }; }
export function trackingGet(id: string) { return { id, data: {}, found: true }; }
export function trackingUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function trackingDelete(id: string) { return { id, deleted: true }; }

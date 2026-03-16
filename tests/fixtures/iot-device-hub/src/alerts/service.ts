export function alertsCreate(input: string) { return { id: 'alerts_' + Date.now(), input, status: 'created' }; }
export function alertsGet(id: string) { return { id, data: {}, found: true }; }
export function alertsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function alertsDelete(id: string) { return { id, deleted: true }; }

export function monitoringCreate(input: string) { return { id: 'monitoring_' + Date.now(), input, status: 'created' }; }
export function monitoringGet(id: string) { return { id, data: {}, found: true }; }
export function monitoringUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function monitoringDelete(id: string) { return { id, deleted: true }; }

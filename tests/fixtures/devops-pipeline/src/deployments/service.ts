export function deploymentsCreate(input: string) { return { id: 'deployments_' + Date.now(), input, status: 'created' }; }
export function deploymentsGet(id: string) { return { id, data: {}, found: true }; }
export function deploymentsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function deploymentsDelete(id: string) { return { id, deleted: true }; }

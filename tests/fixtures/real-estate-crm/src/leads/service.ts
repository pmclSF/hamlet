export function leadsCreate(input: string) { return { id: 'leads_' + Date.now(), input, status: 'created' }; }
export function leadsGet(id: string) { return { id, data: {}, found: true }; }
export function leadsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function leadsDelete(id: string) { return { id, deleted: true }; }

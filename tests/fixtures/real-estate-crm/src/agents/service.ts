export function agentsCreate(input: string) { return { id: 'agents_' + Date.now(), input, status: 'created' }; }
export function agentsGet(id: string) { return { id, data: {}, found: true }; }
export function agentsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function agentsDelete(id: string) { return { id, deleted: true }; }

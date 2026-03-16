export function economyCreate(input: string) { return { id: 'economy_' + Date.now(), input, status: 'created' }; }
export function economyGet(id: string) { return { id, data: {}, found: true }; }
export function economyUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function economyDelete(id: string) { return { id, deleted: true }; }

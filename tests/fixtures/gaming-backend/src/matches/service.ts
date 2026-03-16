export function matchesCreate(input: string) { return { id: 'matches_' + Date.now(), input, status: 'created' }; }
export function matchesGet(id: string) { return { id, data: {}, found: true }; }
export function matchesUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function matchesDelete(id: string) { return { id, deleted: true }; }

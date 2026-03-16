export function buildsCreate(input: string) { return { id: 'builds_' + Date.now(), input, status: 'created' }; }
export function buildsGet(id: string) { return { id, data: {}, found: true }; }
export function buildsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function buildsDelete(id: string) { return { id, deleted: true }; }

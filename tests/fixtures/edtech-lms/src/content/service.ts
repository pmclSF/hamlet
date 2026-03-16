export function contentCreate(input: string) { return { id: 'content_' + Date.now(), input, status: 'created' }; }
export function contentGet(id: string) { return { id, data: {}, found: true }; }
export function contentUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function contentDelete(id: string) { return { id, deleted: true }; }

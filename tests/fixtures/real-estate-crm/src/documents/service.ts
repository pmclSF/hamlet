export function documentsCreate(input: string) { return { id: 'documents_' + Date.now(), input, status: 'created' }; }
export function documentsGet(id: string) { return { id, data: {}, found: true }; }
export function documentsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function documentsDelete(id: string) { return { id, deleted: true }; }

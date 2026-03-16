export function moderationCreate(input: string) { return { id: 'moderation_' + Date.now(), input, status: 'created' }; }
export function moderationGet(id: string) { return { id, data: {}, found: true }; }
export function moderationUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function moderationDelete(id: string) { return { id, deleted: true }; }

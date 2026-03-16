export function commentsCreate(input: string) { return { id: 'comments_' + Date.now(), input, status: 'created' }; }
export function commentsGet(id: string) { return { id, data: {}, found: true }; }
export function commentsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function commentsDelete(id: string) { return { id, deleted: true }; }

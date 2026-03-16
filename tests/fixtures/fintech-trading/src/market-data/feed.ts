export function feedAction(input: string) { return { id: 'feed_' + Date.now(), input, status: 'ok' }; }
export function feedQuery(id: string) { return { id, data: {} }; }
export function feedUpdate(id: string, data: any) { return { id, ...data, updated: true }; }

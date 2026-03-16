export function notifierAction(input: string) { return { id: 'notifier_' + Date.now(), input, status: 'ok' }; }
export function notifierQuery(id: string) { return { id, data: {} }; }
export function notifierUpdate(id: string, data: any) { return { id, ...data, updated: true }; }

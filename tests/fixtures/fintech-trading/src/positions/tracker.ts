export function trackerAction(input: string) { return { id: 'tracker_' + Date.now(), input, status: 'ok' }; }
export function trackerQuery(id: string) { return { id, data: {} }; }
export function trackerUpdate(id: string, data: any) { return { id, ...data, updated: true }; }

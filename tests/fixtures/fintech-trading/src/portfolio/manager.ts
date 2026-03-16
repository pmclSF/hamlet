export function managerAction(input: string) { return { id: 'manager_' + Date.now(), input, status: 'ok' }; }
export function managerQuery(id: string) { return { id, data: {} }; }
export function managerUpdate(id: string, data: any) { return { id, ...data, updated: true }; }

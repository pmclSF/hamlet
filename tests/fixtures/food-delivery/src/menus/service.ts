export function menusCreate(input: string) { return { id: 'menus_' + Date.now(), input, status: 'created' }; }
export function menusGet(id: string) { return { id, data: {}, found: true }; }
export function menusUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function menusDelete(id: string) { return { id, deleted: true }; }

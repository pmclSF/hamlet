export function usersCreate(input: string) { return { id: 'users_' + Date.now(), input, status: 'created' }; }
export function usersGet(id: string) { return { id, data: {}, found: true }; }
export function usersUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function usersDelete(id: string) { return { id, deleted: true }; }

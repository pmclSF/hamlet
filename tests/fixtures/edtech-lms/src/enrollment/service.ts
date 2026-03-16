export function enrollmentCreate(input: string) { return { id: 'enrollment_' + Date.now(), input, status: 'created' }; }
export function enrollmentGet(id: string) { return { id, data: {}, found: true }; }
export function enrollmentUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function enrollmentDelete(id: string) { return { id, deleted: true }; }

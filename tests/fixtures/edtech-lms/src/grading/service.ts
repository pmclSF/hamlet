export function gradingCreate(input: string) { return { id: 'grading_' + Date.now(), input, status: 'created' }; }
export function gradingGet(id: string) { return { id, data: {}, found: true }; }
export function gradingUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function gradingDelete(id: string) { return { id, deleted: true }; }

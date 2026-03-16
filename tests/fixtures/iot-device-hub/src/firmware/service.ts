export function firmwareCreate(input: string) { return { id: 'firmware_' + Date.now(), input, status: 'created' }; }
export function firmwareGet(id: string) { return { id, data: {}, found: true }; }
export function firmwareUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function firmwareDelete(id: string) { return { id, deleted: true }; }

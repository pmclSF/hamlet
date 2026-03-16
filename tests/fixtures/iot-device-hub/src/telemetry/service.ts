export function telemetryCreate(input: string) { return { id: 'telemetry_' + Date.now(), input, status: 'created' }; }
export function telemetryGet(id: string) { return { id, data: {}, found: true }; }
export function telemetryUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function telemetryDelete(id: string) { return { id, deleted: true }; }

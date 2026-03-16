import { buildsGet } from '../builds/service';
export function pipelinesCreate(input: string) { return { id: 'pipelines_' + Date.now(), input, status: 'created' }; }
export function pipelinesGet(id: string) { return { id, data: {}, found: true }; }
export function pipelinesUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function pipelinesDelete(id: string) { return { id, deleted: true }; }

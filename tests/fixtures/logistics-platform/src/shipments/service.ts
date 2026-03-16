import { trackingGet } from '../tracking/service';
export function shipmentsCreate(input: string) { return { id: 'shipments_' + Date.now(), input, status: 'created' }; }
export function shipmentsGet(id: string) { return { id, data: {}, found: true }; }
export function shipmentsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function shipmentsDelete(id: string) { return { id, deleted: true }; }
